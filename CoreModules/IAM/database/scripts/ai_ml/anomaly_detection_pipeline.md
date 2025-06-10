# INNOVABIZ IAM – AI/ML para Detecção de Anomalias

## Objetivo
Implementar uma arquitetura de detecção de anomalias baseada em AI/ML para eventos, acessos, queries e métricas do IAM, integrando com o pipeline de monitoramento já existente.

## 1. Coleta de Dados
- Fonte: Tabelas de auditoria (ex: audit_logs), métricas do Prometheus Exporter, logs de queries (pg_stat_statements), eventos de autenticação.
- Ferramenta sugerida: Python (pandas, scikit-learn, pycaret, prophet, etc), integração via Jupyter ou scripts automatizados.

## 2. Pipeline de Machine Learning
- Pré-processamento: Limpeza, agregação e normalização dos dados.
- Modelos sugeridos:
  - Isolation Forest (sklearn)
  - AutoEncoder (keras/tensorflow)
  - Prophet (previsão temporal)
  - DBSCAN (clusterização de outliers)
- Treinamento: Batch (offline) ou online (streaming, ex: Kafka/Spark).
- Deploy: Exportação do modelo treinado para uso em scripts Python, REST API, ou integração com o Prometheus Alertmanager.

## 3. Exemplo de Script Python (Isolation Forest)
```python
import pandas as pd
from sklearn.ensemble import IsolationForest

# Exemplo: leitura de logs de auditoria
logs = pd.read_csv('audit_logs.csv')
features = logs[['event_type', 'user_id', 'timestamp', 'resource', 'status']]
# Pré-processamento customizado...

model = IsolationForest(contamination=0.01, random_state=42)
model.fit(features)
logs['anomaly_score'] = model.decision_function(features)
logs['is_anomaly'] = model.predict(features) == -1

# Exportar anomalias
anomalies = logs[logs['is_anomaly']]
anomalies.to_csv('anomaly_events.csv', index=False)
```

## 4. Integração e Visualização
- Exportar resultados para dashboards Grafana (via Prometheus Pushgateway, CSV, API REST).
- Gatilhos de alerta automáticos para eventos críticos.
- Relatórios periódicos para compliance e segurança.

## 5. Próximos Passos
- Automatizar coleta e processamento dos dados (cron, Airflow, etc).
- Treinar modelos com dados reais do ambiente IAM.
- Integrar pipeline de detecção ao ciclo CI/CD e monitoramento.
- Avaliar uso de modelos mais sofisticados (LSTM, transformers, etc) conforme maturidade.

---
Dúvidas ou deseja um exemplo de script para outro modelo/linguagem? Solicite!
