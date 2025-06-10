"""
Script para envio automático de alertas de anomalias do IAM para um canal Slack
Salve este arquivo em: scripts/ai_ml/anomaly_alerts_slack.py
"""
import pandas as pd
import requests

# Webhook do Slack (configure no seu workspace)
SLACK_WEBHOOK_URL = 'https://hooks.slack.com/services/SEU/WEBHOOK/URL'
SLACK_CHANNEL = '#canal-alertas'

# Carregar eventos anômalos detectados
anomalies = pd.read_csv('/path/to/anomaly_events.csv')

if not anomalies.empty:
    text = f"*ALERTA IAM: Eventos Anômalos Detectados ({len(anomalies)})*\n"
    text += anomalies.head(10).to_markdown(index=False)
    payload = {
        'channel': SLACK_CHANNEL,
        'username': 'IAM Anomaly Bot',
        'text': text
    }
    response = requests.post(SLACK_WEBHOOK_URL, json=payload)
    if response.status_code == 200:
        print('Alerta Slack enviado com sucesso.')
    else:
        print(f'Erro ao enviar alerta Slack: {response.text}')
else:
    print('Nenhuma anomalia detectada para envio de alerta.')
