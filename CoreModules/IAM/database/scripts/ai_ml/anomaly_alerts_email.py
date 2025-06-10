"""
Script para envio automático de alertas por e-mail a partir de eventos anômalos detectados pelo pipeline AI/ML IAM
Salve este arquivo em: scripts/ai_ml/anomaly_alerts_email.py
"""
import pandas as pd
import smtplib
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart

# Configuração do e-mail
SMTP_SERVER = 'smtp.seuservidor.com'
SMTP_PORT = 587
EMAIL_USER = 'seu_email@dominio.com'
EMAIL_PASS = 'sua_senha'
EMAIL_TO = ['destinatario1@dominio.com', 'destinatario2@dominio.com']

# Carregar eventos anômalos detectados
anomalies = pd.read_csv('/path/to/anomaly_events.csv')

if not anomalies.empty:
    subject = 'ALERTA IAM: Eventos Anômalos Detectados'
    body = anomalies.to_html(index=False)
    msg = MIMEMultipart()
    msg['From'] = EMAIL_USER
    msg['To'] = ', '.join(EMAIL_TO)
    msg['Subject'] = subject
    msg.attach(MIMEText(body, 'html'))

    with smtplib.SMTP(SMTP_SERVER, SMTP_PORT) as server:
        server.starttls()
        server.login(EMAIL_USER, EMAIL_PASS)
        server.sendmail(EMAIL_USER, EMAIL_TO, msg.as_string())
        print('Alerta enviado com sucesso.')
else:
    print('Nenhuma anomalia detectada para envio de alerta.')
