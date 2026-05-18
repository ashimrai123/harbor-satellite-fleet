{{/*
Expand the name of the chart.
*/}}
{{- define "satellite-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "satellite-operator.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "satellite-operator.labels" -}}
helm.sh/chart: {{ include "satellite-operator.name" . }}-{{ .Chart.Version }}
{{ include "satellite-operator.selectorLabels" . }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "satellite-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "satellite-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
