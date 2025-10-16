{{/*
Expand the name of the chart.
*/}}
{{- define "webauthn.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "webauthn.fullname" -}}
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
{{- define "webauthn.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "webauthn.labels" -}}
helm.sh/chart: {{ include "webauthn.chart" . }}
{{ include "webauthn.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: innovabiz-iam
app.kubernetes.io/component: webauthn
{{- end }}

{{/*
Selector labels
*/}}
{{- define "webauthn.selectorLabels" -}}
app.kubernetes.io/name: {{ include "webauthn.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "webauthn.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "webauthn.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the config map
*/}}
{{- define "webauthn.configMapName" -}}
{{- printf "%s-config" (include "webauthn.fullname" .) }}
{{- end }}

{{/*
Create the name of the secret
*/}}
{{- define "webauthn.secretName" -}}
{{- printf "%s-secrets" (include "webauthn.fullname" .) }}
{{- end }}

{{/*
Create database URL from components
*/}}
{{- define "webauthn.databaseUrl" -}}
{{- if .Values.postgresql.enabled }}
{{- printf "postgresql://%s:%s@%s-postgresql:5432/%s" .Values.postgresql.auth.username .Values.postgresql.auth.password .Release.Name .Values.postgresql.auth.database }}
{{- else }}
{{- .Values.secrets.DATABASE_URL }}
{{- end }}
{{- end }}

{{/*
Create Redis URL from components
*/}}
{{- define "webauthn.redisUrl" -}}
{{- if .Values.redis.enabled }}
{{- if .Values.redis.auth.enabled }}
{{- printf "redis://:%s@%s-redis-master:6379" .Values.redis.auth.password .Release.Name }}
{{- else }}
{{- printf "redis://%s-redis-master:6379" .Release.Name }}
{{- end }}
{{- else }}
{{- .Values.secrets.REDIS_URL }}
{{- end }}
{{- end }}

{{/*
Create Kafka brokers from components
*/}}
{{- define "webauthn.kafkaBrokers" -}}
{{- if .Values.kafka.enabled }}
{{- printf "%s-kafka:9092" .Release.Name }}
{{- else }}
{{- .Values.secrets.KAFKA_BROKERS }}
{{- end }}
{{- end }}

{{/*
Generate environment variables for the application
*/}}
{{- define "webauthn.envVars" -}}
- name: DATABASE_URL
  value: {{ include "webauthn.databaseUrl" . | quote }}
- name: REDIS_URL
  value: {{ include "webauthn.redisUrl" . | quote }}
- name: KAFKA_BROKERS
  value: {{ include "webauthn.kafkaBrokers" . | quote }}
{{- range $key, $value := .Values.env }}
- name: {{ $key }}
  value: {{ $value | quote }}
{{- end }}
{{- end }}

{{/*
Generate resource requirements
*/}}
{{- define "webauthn.resources" -}}
{{- if .Values.resources }}
resources:
  {{- if .Values.resources.limits }}
  limits:
    {{- range $key, $value := .Values.resources.limits }}
    {{ $key }}: {{ $value }}
    {{- end }}
  {{- end }}
  {{- if .Values.resources.requests }}
  requests:
    {{- range $key, $value := .Values.resources.requests }}
    {{ $key }}: {{ $value }}
    {{- end }}
  {{- end }}
{{- end }}
{{- end }}

{{/*
Generate security context
*/}}
{{- define "webauthn.securityContext" -}}
{{- if .Values.deployment.securityContext }}
securityContext:
  {{- toYaml .Values.deployment.securityContext | nindent 2 }}
{{- end }}
{{- end }}

{{/*
Generate pod security context
*/}}
{{- define "webauthn.podSecurityContext" -}}
{{- if .Values.deployment.podSecurityContext }}
securityContext:
  {{- toYaml .Values.deployment.podSecurityContext | nindent 2 }}
{{- end }}
{{- end }}

{{/*
Generate image pull secrets
*/}}
{{- define "webauthn.imagePullSecrets" -}}
{{- if .Values.image.pullSecrets }}
imagePullSecrets:
  {{- range .Values.image.pullSecrets }}
  - name: {{ . }}
  {{- end }}
{{- end }}
{{- end }}

{{/*
Generate node selector
*/}}
{{- define "webauthn.nodeSelector" -}}
{{- if .Values.nodeSelector }}
nodeSelector:
  {{- toYaml .Values.nodeSelector | nindent 2 }}
{{- end }}
{{- end }}

{{/*
Generate tolerations
*/}}
{{- define "webauthn.tolerations" -}}
{{- if .Values.tolerations }}
tolerations:
  {{- toYaml .Values.tolerations | nindent 2 }}
{{- end }}
{{- end }}

{{/*
Generate affinity
*/}}
{{- define "webauthn.affinity" -}}
{{- if .Values.affinity }}
affinity:
  {{- toYaml .Values.affinity | nindent 2 }}
{{- end }}
{{- end }}