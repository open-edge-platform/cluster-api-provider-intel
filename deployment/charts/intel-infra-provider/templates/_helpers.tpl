# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

{{/*
Expand the name of the chart.
*/}}
{{- define "intel-infra-provider.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "intel-infra-provider.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "intel-infra-provider.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "intel-infra-provider.labels" -}}
helm.sh/chart: {{ include "intel-infra-provider.chart" . }}
{{ include "intel-infra-provider.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "intel-infra-provider.selectorLabels" -}}
app.kubernetes.io/name: {{ include "intel-infra-provider.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
manager labels
*/}}
{{- define "intel-infra-provider.managerLabels" -}}
{{ include "intel-infra-provider.selectorLabels" . }}
{{- with .Values.manager.labels }}
{{- toYaml . }}
{{- end }}
{{- end }}

{{/*
manager pod labels
*/}}
{{- define "intel-infra-provider.managerPodLabels" -}}
{{ include "intel-infra-provider.selectorLabels" . }}
{{- with .Values.manager.podLabels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
southbound api labels
*/}}
{{- define "intel-infra-provider.southboundLabels" -}}
{{ include "intel-infra-provider.selectorLabels" . }}
{{- with .Values.southboundApi.labels }}
{{- toYaml . }}
{{- end }}
{{- end }}

{{/*
southbound api pod labels
*/}}
{{- define "intel-infra-provider.southboundPodLabels" -}}
{{ include "intel-infra-provider.selectorLabels" . }}
{{- with .Values.southboundApi.podLabels }}
{{- toYaml . }}
{{- end }}
{{- end }}

{{/*
Role labels
*/}}
{{- define "intel-infra-provider.roleLabels" -}}
{{ include "intel-infra-provider.selectorLabels" . }}
{{- with .Values.rbac.labels }}
{{- toYaml . }}
{{- end }}
{{- end }}

{{/*
Service account labels
*/}}
{{- define "intel-infra-provider.serviceAccountLabels" -}}
{{ include "intel-infra-provider.selectorLabels" . }}
{{- with .Values.rbac.serviceAccount.labels }}
{{- toYaml . }}
{{- end }}
{{- end }}

{{/*
Service monitor labels
*/}}
{{- define "intel-infra-provider.serviceMonitorLabels" -}}
{{ include "intel-infra-provider.selectorLabels" . }}
{{- with .Values.metrics.serviceMonitor.labels }}
{{- toYaml . }}
{{- end }}
{{- end }}

{{/*
Network Policy labels
*/}}
{{- define "intel-infra-provider.networkPolicyLabels" -}}
{{ include "intel-infra-provider.selectorLabels" . }}
{{- with .Values.metrics.networkPolicy.labels }}
{{- toYaml . }}
{{- end }}
{{- end }}

{{/*
Metrics service labels
*/}}
{{- define "intel-infra-provider.metricsServiceLabels" -}}
{{ include "intel-infra-provider.selectorLabels" . }}
{{- with .Values.metrics.service.labels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "intel-infra-provider.serviceAccountName" -}}
{{- if .Values.rbac.serviceAccount.create }}
{{- default (include "intel-infra-provider.fullname" .) .Values.rbac.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.rbac.serviceAccount.name }}
{{- end }}
{{- end }}
