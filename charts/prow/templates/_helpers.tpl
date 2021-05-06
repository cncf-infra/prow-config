{{/*
Expand the name of the chart.
*/}}
{{- define "prow.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "prow.fullname" -}}
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
{{- define "prow.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "prow.labels" -}}
helm.sh/chart: {{ include "prow.chart" . }}
{{ include "prow.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "prow.selectorLabels" -}}
app.kubernetes.io/name: {{ include "prow.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "prow.configmap.volume" -}}
- name: config
  configMap:
    {{- if .Values.configFromConfigMap.enabled }}
    name: {{ .Values.configFromConfigMap.name }}
    {{- else }}
    name: {{ include "prow.fullname" . }}-config
    {{- end }}
{{- end }}

{{- define "prow.plugins.volume" -}}
- name: plugins
  configMap:
    {{- if .Values.pluginsFromConfigMap.enabled }}
    name: {{ .Values.pluginsFromConfigMap.name }}
    {{- else }}
    name: {{ include "prow.fullname" . }}-plugins
    {{- end }}
{{- end }}

{{- define "prow.github-token.volume" -}}
- name: github-secrets-token
  secret:
    defaultMode: 420
    {{- if .Values.githubFromSecretRef.enabled }}
    secretName: {{ .Values.githubFromSecretRef.oauth.name }}
    {{- else }}
    secretName: {{ include "prow.fullname" . }}-github-secrets-token
    {{- end }}
{{- end }}

{{- define "prow.github-hmac.volume" -}}
- name: github-secrets-hmac
  secret:
    defaultMode: 420
    {{- if .Values.githubFromSecretRef.enabled }}
    secretName: {{ .Values.githubFromSecretRef.hmac.name }}
    {{- else }}
    secretName: {{ include "prow.fullname" . }}-github-secrets-hmac
    {{- end }}
{{- end }}

{{- define "prow.github-cookie.volume" -}}
- name: github-secrets-cookie
  secret:
    defaultMode: 420
    {{- if .Values.githubFromSecretRef.enabled }}
    secretName: {{ .Values.githubFromSecretRef.cookie.name }}
    {{- else }}
    secretName: {{ include "prow.fullname" . }}-github-secrets-cookie
    {{- end }}
{{- end }}
