"""
Script de teste automatizado do pipeline de detecção de anomalias IAM
Salve em: scripts/ai_ml/test_pipeline.py
"""
import pandas as pd
import subprocess
import os

# 1. Gerar arquivo de eventos anômalos simulado
test_data = [
    {'event_type': 'login', 'user_id': 123, 'timestamp': '2025-06-10T20:00:00Z', 'resource': 'api/v1/auth', 'status': 'failed'},
    {'event_type': 'delete', 'user_id': 999, 'timestamp': '2025-06-10T20:01:00Z', 'resource': 'api/v1/user', 'status': 'success'},
    {'event_type': 'update', 'user_id': 888, 'timestamp': '2025-06-10T20:02:00Z', 'resource': 'api/v1/role', 'status': 'failed'}
]
df = pd.DataFrame(test_data)
df.to_csv('anomaly_events.csv', index=False)
print('[TEST] Arquivo anomaly_events.csv gerado.')

# 2. Executar scripts de alerta (ajuste paths conforme necessário)
print('[TEST] Executando alerta por e-mail...')
subprocess.run(['python', 'anomaly_alerts_email.py'])

print('[TEST] Executando alerta por Slack...')
subprocess.run(['python', 'anomaly_alerts_slack.py'])

print('[TEST] Pipeline de teste concluído. Verifique os canais de alerta.')
