"""
DAG Airflow para detecção automática de anomalias no IAM
Salve este arquivo em: scripts/ai_ml/anomaly_detection_airflow_dag.py
"""
from airflow import DAG
from airflow.operators.python_operator import PythonOperator
from datetime import datetime, timedelta
import pandas as pd
from sklearn.ensemble import IsolationForest

def detect_anomalies():
    # Exemplo: leitura dos logs exportados do IAM
    logs = pd.read_csv('/path/to/audit_logs.csv')
    features = logs[['event_type', 'user_id', 'timestamp', 'resource', 'status']]
    # Pré-processamento customizado...
    model = IsolationForest(contamination=0.01, random_state=42)
    model.fit(features)
    logs['anomaly_score'] = model.decision_function(features)
    logs['is_anomaly'] = model.predict(features) == -1
    anomalies = logs[logs['is_anomaly']]
    anomalies.to_csv('/path/to/anomaly_events.csv', index=False)
    # Aqui pode-se adicionar integração com alertas, API, etc

default_args = {
    'owner': 'airflow',
    'depends_on_past': False,
    'start_date': datetime(2025, 6, 10),
    'retries': 1,
    'retry_delay': timedelta(minutes=5)
}

dag = DAG(
    'iam_anomaly_detection',
    default_args=default_args,
    schedule_interval='@daily',
    catchup=False
)

task_detect_anomalies = PythonOperator(
    task_id='detect_anomalies',
    python_callable=detect_anomalies,
    dag=dag
)
