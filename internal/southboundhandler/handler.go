// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package southboundhandler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	client "k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	cutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	infrastructurev1alpha1 "github.com/open-edge-platform/cluster-api-provider-intel/api/v1alpha1"
	pb "github.com/open-edge-platform/cluster-api-provider-intel/pkg/api/proto"
	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/logging"
	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/tenant"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	cloudinit "sigs.k8s.io/cluster-api/test/infrastructure/docker/cloudinit"
)

const (
	maxLabelValLen = 63
)

var (
	log                        = logging.GetLogger("handler")
	IntelMachineResourceSchema = schema.GroupVersionResource{Group: infrastructurev1alpha1.GroupVersion.Group, Version: infrastructurev1alpha1.GroupVersion.Version, Resource: "intelmachines"}
	MachineResourceSchema      = schema.GroupVersionResource{Group: clusterv1.GroupVersion.Group, Version: clusterv1.GroupVersion.Version, Resource: "machines"}
	alphaNum                   = regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString
	labelVal                   = regexp.MustCompile(`^[a-zA-Z0-9]+[a-zA-Z0-9-_.]*[a-zA-Z0-9]+$`).MatchString
)

const (
	rateLimiterQPS   = "RATE_LIMITER_QPS"
	rateLimiterBurst = "RATE_LIMITER_BURST"
	defaultQPS       = 30
	defaultBurst     = 100
)

// validLabelVal checks if the string argument is a valid k8s label value
// must be 63 characters or less (can be empty),
// unless empty, must begin and end with an alphanumeric character ([a-z0-9A-Z]),
// could contain dashes (-), underscores (_), dots (.), and alphanumerics between.
func validLabelVal(val string) bool {
	if len(val) == 0 ||
		len(val) <= 2 && alphaNum(val) ||
		len(val) <= maxLabelValLen && labelVal(val) {
		return true
	}
	return false
}

type Handler struct {
	client client.Client
}

func NewHandler() (*Handler, error) {
	// Create the in-cluster config
	config, err := client.InClusterConfig()
	if err != nil {
		return nil, err
	}

	// Set rate limiter parameters
	qpsValue, burstValue, err := getRateLimiterParams()
	if err != nil {
		log.Warn().Err(err).Msg("unable to get rate limiter params; using default values")
	}
	log.Info().Msgf("rate limiter params: qps: %v, burst: %v", qpsValue, burstValue)

	config.QPS = float32(qpsValue)
	config.Burst = int(burstValue)

	// Create a controller-runtime manager
	mgr, err := ctrl.NewManager(config, ctrl.Options{
		Scheme: runtime.NewScheme(), // Add your scheme here if needed
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create manager: %w", err)
	}

	// Use the manager's cached client
	cachedClient := mgr.GetClient()

	// Start the manager in a separate goroutine
	go func() {
		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			log.Fatal().Err(err).Msg("failed to start manager")
		}
	}()

	return &Handler{client: cachedClient}, nil
}

// Register is called by the CO Agent when registring a new cluster node.
// It retrieves the IntelMachine and Machine resources, extracts the bootstrap script from the secret,
// and returns the install and uninstall commands.
//
// Parameters:
// - ctx: The context for the request.
// - nodeGUID: The GUID of the node to be registered.
//
// Returns:
// - installCmd: The command to install the node.
// - uninstallCmd: The command to uninstall the node.
// - result: The result of the registration process.
// - err: Any error encountered during the registration process.
func (h *Handler) Register(ctx context.Context, nodeGUID string) (*pb.ShellScriptCommand, *pb.ShellScriptCommand, pb.RegisterClusterResponse_Result, error) {
	log.Info().Msgf("Registering node %s\n", nodeGUID)

	// Get Project ID from context
	projectId := tenant.GetActiveProjectIdFromContext(ctx)

	// Get IntelMachine in namespace <Project ID> with matching nodeGUID
	intelmachine, err := getIntelMachine(ctx, h.client, projectId, nodeGUID)
	if err != nil {
		log.Error().Msg("Failed to get IntelMachine")
		return nil, nil, pb.RegisterClusterResponse_ERROR, err
	}
	if intelmachine == nil {
		err := errors.New("IntelMachine not found")
		log.Error().Msg(err.Error())
		return nil, nil, pb.RegisterClusterResponse_ERROR, err
	}
	providerID := intelmachine.Spec.ProviderID
	if providerID == nil {
		err := errors.New("IntelMachine does not have a ProviderID")
		log.Error().Msg(err.Error())
		return nil, nil, pb.RegisterClusterResponse_ERROR, err
	}

	ownername, err := getMachineOwnerName(intelmachine)
	if err != nil {
		log.Error().Msg("Failed to find owner reference on IntelMachine")
		return nil, nil, pb.RegisterClusterResponse_ERROR, err
	}
	machine, err := getMachine(ctx, h.client, projectId, ownername)
	if err != nil {
		log.Error().Msg("Failed to get Machine")
		return nil, nil, pb.RegisterClusterResponse_ERROR, err
	}

	// Get the Secret that matches Machine.Spec.Bootstrap.DataSecretName
	if machine.Spec.Bootstrap.DataSecretName == nil {
		err := errors.New("machine does not have a DataSecretName")
		log.Error().Msg(err.Error())
		return nil, nil, pb.RegisterClusterResponse_ERROR, err
	}
	secretName := *machine.Spec.Bootstrap.DataSecretName
	secret, err := getSecret(ctx, h.client, projectId, secretName)
	if err != nil {
		log.Error().Msg("Failed to get Bootstrap Secret")
		return nil, nil, pb.RegisterClusterResponse_ERROR, err
	}

	kind := machine.Spec.Bootstrap.ConfigRef.Kind

	// Convert the Secret to the install command
	install, err := extractBootstrapScript(secret, kind, *providerID)
	if err != nil {
		log.Error().Msgf("Failed to extract Bootstrap Secret: %s", err)
		return nil, nil, pb.RegisterClusterResponse_ERROR, err
	}

	uninstall, err := getUninstall(kind)
	if err != nil {
		log.Error().Msgf("Failed to get uninstall command: %s", err)
		return nil, nil, pb.RegisterClusterResponse_ERROR, err
	}

	installCmd := &pb.ShellScriptCommand{Command: install}
	uninstallCmd := &pb.ShellScriptCommand{Command: uninstall}
	return installCmd, uninstallCmd, pb.RegisterClusterResponse_SUCCESS, nil
}

// UpdateStatus determines the action that the CO Agent should take next based on its
// current state.  It saves the CO Agent's current state in an annotation on the IntelMachine.
//
// Parameters:
//   - ctx: The context for the request.
//   - nodeGUID: The unique identifier of the node.
//   - status: The status code from the CO Agent.
//
// Returns:
//   - pb.UpdateClusterStatusResponse_ActionRequest: The action request to be taken based on the status update.
//   - error: An error if the status update fails.
func (h *Handler) UpdateStatus(ctx context.Context, nodeGUID string, status pb.UpdateClusterStatusRequest_Code) (pb.UpdateClusterStatusResponse_ActionRequest, error) {
	var hostState string

	// Default action is NONE
	action := pb.UpdateClusterStatusResponse_NONE

	// Get Project ID from context
	projectId := tenant.GetActiveProjectIdFromContext(ctx)

	// Get IntelMachine in namespace <Project ID> with matching nodeGUID
	intelmachine, err := getIntelMachine(ctx, h.client, projectId, nodeGUID)
	if err != nil {
		log.Error().Msg("Failed to get IntelMachine")
		return pb.UpdateClusterStatusResponse_NONE, err
	}
	if intelmachine == nil {
		// The node has not yet been put into a cluster
		return pb.UpdateClusterStatusResponse_NONE, nil
	}

	currentHostState := intelmachine.Annotations[infrastructurev1alpha1.HostStateAnnotation]
	removeFinalizer := false

	// Choose appropriate ActionRequest
	switch status {
	case pb.UpdateClusterStatusRequest_INACTIVE:
		hostState = infrastructurev1alpha1.HostStateInactive

		// If IntelMachine is not deleted and has a ProviderID, it's time to bootstrap the node
		if intelmachine.DeletionTimestamp.IsZero() {
			if intelmachine.Spec.ProviderID != nil {
				action = pb.UpdateClusterStatusResponse_REGISTER
			}
		} else {
			if cutil.ContainsFinalizer(intelmachine, infrastructurev1alpha1.HostCleanupFinalizer) {
				removeFinalizer = true
			}
		}

	case pb.UpdateClusterStatusRequest_REGISTERING, pb.UpdateClusterStatusRequest_INSTALL_IN_PROGRESS:
		hostState = infrastructurev1alpha1.HostStateInProgress

	case pb.UpdateClusterStatusRequest_ACTIVE:
		hostState = infrastructurev1alpha1.HostStateActive

		// If IntelMachine is being deleted, need to clean up the node
		if !intelmachine.DeletionTimestamp.IsZero() {
			action = pb.UpdateClusterStatusResponse_DEREGISTER
		}

	case pb.UpdateClusterStatusRequest_DEREGISTERING, pb.UpdateClusterStatusRequest_UNINSTALL_IN_PROGRESS:
		hostState = infrastructurev1alpha1.HostStateInProgress

	case pb.UpdateClusterStatusRequest_ERROR:
		hostState = infrastructurev1alpha1.HostStateError
	}

	// Only patch IntelMachine if it needs to be changed
	if currentHostState != hostState || removeFinalizer {
		origIntelMachine := intelmachine.DeepCopy()

		if removeFinalizer {
			cutil.RemoveFinalizer(intelmachine, infrastructurev1alpha1.HostCleanupFinalizer)
		}

		// Update the IntelMachine annotations
		if intelmachine.Annotations == nil {
			intelmachine.Annotations = make(map[string]string)
		}
		intelmachine.Annotations[infrastructurev1alpha1.HostStateAnnotation] = hostState
		return action, patchIntelMachine(ctx, h.client, origIntelMachine, intelmachine)
	}

	return action, nil
}

func getIntelMachine(ctx context.Context, client dynamic.Interface, projectId string, nodeGUID string) (*infrastructurev1alpha1.IntelMachine, error) {
	if !validLabelVal(nodeGUID) {
		return nil, errors.New("invalid Node GUID")
	}
	opts := v1.ListOptions{
		LabelSelector: infrastructurev1alpha1.NodeGUIDKey + "=" + nodeGUID,
	}
	result, err := client.Resource(IntelMachineResourceSchema).Namespace(projectId).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	if len(result.Items) < 1 {
		return nil, nil
	}
	if len(result.Items) > 1 {
		return nil, errors.New("duplicate IntelMachines found")
	}
	intelmachine := &infrastructurev1alpha1.IntelMachine{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(result.Items[0].Object, intelmachine)
	if err != nil {
		return nil, err
	}
	if intelmachine.Spec.NodeGUID != nodeGUID {
		return nil, errors.New("invalid IntelMachine found")
	}
	return intelmachine, nil
}

func patchIntelMachine(ctx context.Context, client dynamic.Interface, orig, new *infrastructurev1alpha1.IntelMachine) error {
	patchBytes, err := getPatchData(orig, new)
	if err != nil {
		return err
	}

	_, err = client.Resource(IntelMachineResourceSchema).Namespace(orig.Namespace).Patch(ctx, orig.GetName(), types.MergePatchType, patchBytes, v1.PatchOptions{})
	if err != nil {
		return err
	}

	return nil
}

func getMachine(ctx context.Context, client dynamic.Interface, projectId string, name string) (*clusterv1.Machine, error) {
	result, err := client.Resource(MachineResourceSchema).Namespace(projectId).Get(ctx, name, v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	machine := &clusterv1.Machine{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(result.Object, machine)
	if err != nil {
		return nil, err
	}

	return machine, nil
}

func getSecret(ctx context.Context, client dynamic.Interface, projectId string, name string) (*corev1.Secret, error) {
	result, err := client.Resource(schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}).Namespace(projectId).Get(ctx, name, v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	secret := &corev1.Secret{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(result.Object, secret)
	if err != nil {
		return nil, err
	}

	return secret, nil
}

func getMachineOwnerName(intelmachine *infrastructurev1alpha1.IntelMachine) (string, error) {
	ownerrefs := intelmachine.GetOwnerReferences()
	for _, ownerref := range ownerrefs {
		if ownerref.Kind == "Machine" {
			return ownerref.Name, nil
		}
	}
	return "", errors.New("machine not found")
}

func extractBootstrapScript(secret *corev1.Secret, kind, providerID string) (string, error) {
	format, ok := secret.Data["format"]
	if !ok {
		return "", errors.New("missing format in bootstrap secret")
	}

	if string(format) != "cloud-config" {
		return "", errors.New("unsupported bootstrap script format: " + string(format))
	}

	value, ok := secret.Data["value"]
	if !ok {
		return "", errors.New("missing value in bootstrap secret")
	}

	commands, err := cloudinit.Commands(value)
	if err != nil {
		return "", err
	}
	switch {
	case kind == "KubeadmConfig":
		// Add providerID to Kubeadm node
	case kind == "RKE2Config":
		dir := "/etc/rancher/rke2/config.yaml.d/"
		filename := dir + "providerID.yaml"
		newcmds := []cloudinit.Cmd{
			{
				Cmd:  "mkdir",
				Args: []string{"-p", dir},
			},
			{
				Cmd:   "/bin/sh",
				Args:  []string{"-c", fmt.Sprintf("cat > %s /dev/stdin", filename)},
				Stdin: fmt.Sprintf(`kubelet-arg+: [\"--provider-id=%s\"]`, providerID),
			},
			{
				Cmd:  "chmod",
				Args: []string{"0640", filename},
			},
		}
		commands = append(newcmds, commands...)
	default:
		return "", fmt.Errorf("unsupported bootstrap provider: %s", kind)
	}

	shellcommands := make([]string, len(commands))
	for i, cmd := range commands {
		shellcommand, err := getCommand(cmd)
		if err != nil {
			return "", err
		}
		shellcommands[i] = shellcommand
	}
	script := strings.Join(shellcommands, "; ")
	script = fmt.Sprintf("sudo sh -c \"%s\"", script)

	return script, nil
}

// Determine the uninstall command from the bootstrap kind
func getUninstall(kind string) (string, error) {
	uninstall := ""
	var err error = nil
	if kind == "KubeadmConfig" {
		uninstall = "sudo /usr/local/bin/kubeadm-uninstall.sh"
	} else if kind == "RKE2Config" {
		uninstall = "if [ -f /usr/local/bin/rke2-uninstall.sh ]; then sudo /usr/local/bin/rke2-uninstall.sh; else sudo /opt/rke2/bin/rke2-uninstall.sh; fi"
	} else {
		err = fmt.Errorf("unknown bootstrap provider: %s", kind)
	}
	return uninstall, err
}

// getPatchData will return difference between original and modified document
func getPatchData(originalObj, modifiedObj interface{}) ([]byte, error) {
	originalData, err := json.Marshal(originalObj)
	if err != nil {
		return nil, err
	}
	modifiedData, err := json.Marshal(modifiedObj)
	if err != nil {
		return nil, err
	}

	patchBytes, err := jsonpatch.CreateMergePatch(originalData, modifiedData)
	if err != nil {
		return nil, err
	}
	return patchBytes, nil
}

// Convert a cloudinit.Cmd to a runnable shell command.
// This is bare-bones at present but seems adequate for the RKE2 Bootstrap script.
func getCommand(cmd cloudinit.Cmd) (string, error) {
	switch {
	case cmd.Cmd == "mkdir" || cmd.Cmd == "chmod":
		return fmt.Sprintf("%s %s", cmd.Cmd, strings.Join(cmd.Args, " ")), nil
	case cmd.Cmd == "/bin/sh":
		if len(cmd.Args) == 2 {
			if cmd.Args[0] == "-c" {
				if cmd.Stdin != "" {
					args := strings.Split(cmd.Args[1], " ")
					if len(args) == 4 && args[0] == "cat" && args[1] == ">" && args[3] == "/dev/stdin" {
						return fmt.Sprintf("echo '%s' > %s", cmd.Stdin, args[2]), nil
					}
				} else {
					return cmd.Args[1], nil
				}
			}
		}
	}

	return "", fmt.Errorf("unparsed command: %+v", cmd)
}

func getRateLimiterParams() (float64, int64, error) {
	qps := os.Getenv(rateLimiterQPS)
	qpsValue, err := strconv.ParseFloat(qps, 32)
	if err != nil {
		return defaultQPS, defaultBurst, err
	}
	burst := os.Getenv(rateLimiterBurst)
	burstValue, err := strconv.ParseInt(burst, 10, 32)
	if err != nil {
		return defaultQPS, defaultBurst, err
	}
	return qpsValue, burstValue, nil
}
