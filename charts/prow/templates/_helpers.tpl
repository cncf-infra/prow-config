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

{{- define "prow.github-oauth-config.volume" -}}
- name: github-oauth-config
  secret:
    defaultMode: 420
    {{- if .Values.githubFromSecretRef.enabled }}
    secretName: {{ .Values.githubFromSecretRef.oauthConfig.name }}
    {{- else }}
    secretName: {{ include "prow.fullname" . }}-github-oauth-config
    {{- end }}
{{- end }}

{{- define "prow.hook-setup" -}}
metadata:
{{- with .Values.hook.podAnnotations }}
  annotations:
    {{- toYaml . | nindent 4 }}
{{- end }}
  labels:
    app.kubernetes.io/component: hook
    {{- include "prow.selectorLabels" . | nindent 4 }}
spec:
  restartPolicy: OnFailure
  {{- with .Values.imagePullSecrets }}
  imagePullSecrets:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  serviceAccountName: {{ include "prow.fullname" . }}-hook-setupjob
  containers:
  - name: {{ .Chart.Name }}-hook-setupjob
    image: "{{ .Values.hook.setupJob.image.repository }}:{{ .Values.hook.setupJob.image.tag | default .Chart.AppVersion }}"
    imagePullPolicy: {{ .Values.hook.setupJob.image.pullPolicy }}
    securityContext:
      {{- toYaml .Values.hook.setupJob.podSecurityContext | nindent 6 }}
    command:
    - /hmac
    args:
    - --config-path=/etc/config/config.yaml
    {{ range $index, $host := .Values.ingress.hosts }}
    - --hook-url=https://{{ $host.host }}/hook
    {{- end }}
    {{- if .Values.githubFromSecretRef.enabled }}
    - --hmac-token-secret-name={{ .Values.githubFromSecretRef.hmac.name }}
    {{- else }}
    - --hmac-token-secret-name={{ include "prow.fullname" . }}-github-secrets-hmac
    {{- end }}
    - --hmac-token-secret-namespace={{ .Release.Namespace }}
    - --hmac-token-key=hmac
    - --github-token-path=/etc/github/oauth
    - --github-endpoint=http://ghproxy.{{ .Release.Namespace }}
    - --github-endpoint=https://api.github.com
    - --kubeconfig-context=default
    - --dry-run=false
    volumeMounts:
      - name: github-secrets-token
        mountPath: /etc/github/oauth
        subPath: oauth
        readOnly: true
      - name: github-secrets-hmac
        mountPath: /etc/github/hmac
        subPath: hmac
        readOnly: true
      - name: config
        mountPath: /etc/config
        readOnly: true
  {{- with .Values.hook.nodeSelector }}
  nodeSelector:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  {{- with .Values.hook.affinity }}
  affinity:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  {{- with .Values.hook.tolerations }}
  tolerations:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  volumes:
  {{- include "prow.github-token.volume" . | nindent 4 }}
  {{- include "prow.github-hmac.volume" . | nindent 4 }}
  {{- include "prow.configmap.volume" . | nindent 4 }}
{{- end }}
