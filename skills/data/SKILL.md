---
name: data
description: >
  Use when: "data pipeline", "ETL", "Spark", "dbt", "Airflow", "data warehouse",
  "analytics", "machine learning", "ML", "model", "PyTorch", "TensorFlow",
  "MLOps", "MLflow", "Kubeflow", "feature engineering", "A/B testing".
version: 1.0.0
author: "AgentOps Team"
license: "MIT"
---

# Data Skill

Data engineering, data science, ML engineering, and MLOps patterns.

## Quick Reference

| Area | Key Patterns | When to Use |
|------|--------------|-------------|
| **Data Engineering** | Pipelines, Spark, dbt, Airflow | Data infrastructure |
| **Data Science** | Analytics, ML, statistics | Analysis & modeling |
| **ML Engineering** | PyTorch, TensorFlow, serving | Production ML |
| **MLOps** | MLflow, Kubeflow, pipelines | ML lifecycle |

---

## Data Engineering

### Pipeline Architecture

```
[Sources] → [Ingestion] → [Transform] → [Storage] → [Serving]
   │            │             │            │            │
   │            │             │            │            └─ APIs, Dashboards
   │            │             │            └─ Data Warehouse
   │            │             └─ Spark, dbt
   │            └─ Kafka, Airbyte
   └─ Databases, APIs, Files
```

### ETL vs ELT

| Pattern | When to Use |
|---------|-------------|
| **ETL** | Transform before loading (legacy, complex transforms) |
| **ELT** | Load then transform (modern warehouses, dbt) |

### Spark Patterns

```python
from pyspark.sql import SparkSession
from pyspark.sql.functions import col, when, sum as spark_sum

spark = SparkSession.builder.appName("pipeline").getOrCreate()

# Read
df = spark.read.parquet("s3://bucket/data/")

# Transform
result = (df
    .filter(col("status") == "active")
    .groupBy("category")
    .agg(spark_sum("amount").alias("total"))
    .orderBy(col("total").desc())
)

# Write
result.write.mode("overwrite").parquet("s3://bucket/output/")
```

### dbt Patterns

```sql
-- models/staging/stg_orders.sql
{{ config(materialized='view') }}

select
    id as order_id,
    customer_id,
    order_date,
    status,
    total_amount
from {{ source('raw', 'orders') }}
where status != 'cancelled'

-- models/marts/fct_daily_revenue.sql
{{ config(materialized='table') }}

select
    date_trunc('day', order_date) as date,
    count(*) as order_count,
    sum(total_amount) as revenue
from {{ ref('stg_orders') }}
group by 1
```

### Airflow DAG

```python
from airflow import DAG
from airflow.operators.python import PythonOperator
from datetime import datetime, timedelta

default_args = {
    'retries': 3,
    'retry_delay': timedelta(minutes=5),
}

with DAG(
    'daily_pipeline',
    default_args=default_args,
    schedule_interval='0 6 * * *',
    start_date=datetime(2024, 1, 1),
    catchup=False,
) as dag:

    extract = PythonOperator(
        task_id='extract',
        python_callable=extract_data,
    )

    transform = PythonOperator(
        task_id='transform',
        python_callable=transform_data,
    )

    load = PythonOperator(
        task_id='load',
        python_callable=load_data,
    )

    extract >> transform >> load
```

---

## Data Science

### Analysis Workflow

```python
import pandas as pd
import numpy as np
from sklearn.model_selection import train_test_split
from sklearn.preprocessing import StandardScaler

# Load and explore
df = pd.read_csv("data.csv")
print(df.describe())
print(df.info())
print(df.isnull().sum())

# Clean
df = df.dropna(subset=['target'])
df['feature'] = df['feature'].fillna(df['feature'].median())

# Feature engineering
df['feature_log'] = np.log1p(df['feature'])
df['feature_squared'] = df['feature'] ** 2

# Split
X = df.drop('target', axis=1)
y = df['target']
X_train, X_test, y_train, y_test = train_test_split(
    X, y, test_size=0.2, random_state=42
)

# Scale
scaler = StandardScaler()
X_train_scaled = scaler.fit_transform(X_train)
X_test_scaled = scaler.transform(X_test)
```

### Statistical Testing

```python
from scipy import stats

# T-test (comparing means)
t_stat, p_value = stats.ttest_ind(group_a, group_b)

# Chi-square (categorical association)
chi2, p_value, dof, expected = stats.chi2_contingency(contingency_table)

# A/B test significance
from statsmodels.stats.proportion import proportions_ztest
z_stat, p_value = proportions_ztest(
    count=[conversions_a, conversions_b],
    nobs=[visitors_a, visitors_b],
)
```

### ML Model Training

```python
from sklearn.ensemble import RandomForestClassifier
from sklearn.metrics import classification_report, roc_auc_score

# Train
model = RandomForestClassifier(n_estimators=100, random_state=42)
model.fit(X_train, y_train)

# Evaluate
y_pred = model.predict(X_test)
y_prob = model.predict_proba(X_test)[:, 1]

print(classification_report(y_test, y_pred))
print(f"ROC-AUC: {roc_auc_score(y_test, y_prob):.3f}")
```

---

## ML Engineering

### PyTorch Model

```python
import torch
import torch.nn as nn
import torch.optim as optim

class Model(nn.Module):
    def __init__(self, input_dim, hidden_dim, output_dim):
        super().__init__()
        self.fc1 = nn.Linear(input_dim, hidden_dim)
        self.fc2 = nn.Linear(hidden_dim, output_dim)
        self.relu = nn.ReLU()
        self.dropout = nn.Dropout(0.2)

    def forward(self, x):
        x = self.relu(self.fc1(x))
        x = self.dropout(x)
        x = self.fc2(x)
        return x

# Training loop
model = Model(input_dim=10, hidden_dim=64, output_dim=2)
optimizer = optim.Adam(model.parameters(), lr=0.001)
criterion = nn.CrossEntropyLoss()

for epoch in range(100):
    model.train()
    optimizer.zero_grad()
    outputs = model(X_train_tensor)
    loss = criterion(outputs, y_train_tensor)
    loss.backward()
    optimizer.step()
```

### Model Serving

```python
# FastAPI model serving
from fastapi import FastAPI
import torch

app = FastAPI()
model = torch.load("model.pt")
model.eval()

@app.post("/predict")
async def predict(features: list[float]):
    with torch.no_grad():
        tensor = torch.tensor([features])
        prediction = model(tensor)
        return {"prediction": prediction.tolist()}
```

### Feature Engineering

```python
# Feature store pattern
from datetime import datetime, timedelta

def compute_user_features(user_id: str, as_of: datetime) -> dict:
    """Compute features for a user as of a given time."""
    return {
        "user_id": user_id,
        "orders_30d": count_orders(user_id, as_of - timedelta(days=30), as_of),
        "total_spend_30d": sum_spend(user_id, as_of - timedelta(days=30), as_of),
        "avg_order_value": avg_order_value(user_id, as_of),
        "days_since_last_order": days_since_last(user_id, as_of),
    }
```

---

## MLOps

### MLflow Tracking

```python
import mlflow
import mlflow.sklearn

mlflow.set_experiment("my-experiment")

with mlflow.start_run():
    # Log parameters
    mlflow.log_param("n_estimators", 100)
    mlflow.log_param("max_depth", 10)

    # Train model
    model = train_model(params)

    # Log metrics
    mlflow.log_metric("accuracy", accuracy)
    mlflow.log_metric("f1_score", f1)

    # Log model
    mlflow.sklearn.log_model(model, "model")

    # Log artifacts
    mlflow.log_artifact("confusion_matrix.png")
```

### ML Pipeline (Kubeflow)

```python
from kfp import dsl
from kfp.components import func_to_container_op

@func_to_container_op
def preprocess(input_path: str) -> str:
    # Preprocessing logic
    return output_path

@func_to_container_op
def train(data_path: str, epochs: int) -> str:
    # Training logic
    return model_path

@func_to_container_op
def evaluate(model_path: str, test_path: str) -> float:
    # Evaluation logic
    return accuracy

@dsl.pipeline(name="ML Pipeline")
def ml_pipeline(input_data: str, epochs: int = 10):
    preprocess_task = preprocess(input_data)
    train_task = train(preprocess_task.output, epochs)
    evaluate_task = evaluate(train_task.output, preprocess_task.output)
```

### Model Registry

```python
# Register model with MLflow
from mlflow.tracking import MlflowClient

client = MlflowClient()

# Register model
model_uri = f"runs:/{run_id}/model"
model_version = mlflow.register_model(model_uri, "my-model")

# Transition to production
client.transition_model_version_stage(
    name="my-model",
    version=model_version.version,
    stage="Production",
)
```

### A/B Testing for Models

```python
import random

class ModelRouter:
    def __init__(self, models: dict, weights: dict):
        self.models = models
        self.weights = weights

    def predict(self, features, user_id: str):
        # Deterministic assignment based on user_id
        assignment = hash(user_id) % 100

        cumulative = 0
        for model_name, weight in self.weights.items():
            cumulative += weight
            if assignment < cumulative:
                model = self.models[model_name]
                prediction = model.predict(features)
                return {
                    "prediction": prediction,
                    "model": model_name,
                }

# Usage
router = ModelRouter(
    models={"control": model_v1, "treatment": model_v2},
    weights={"control": 50, "treatment": 50},
)
```

---

## Data Quality

### Data Validation

```python
import great_expectations as ge

# Create expectation suite
df = ge.from_pandas(pandas_df)

# Add expectations
df.expect_column_to_exist("user_id")
df.expect_column_values_to_not_be_null("user_id")
df.expect_column_values_to_be_unique("user_id")
df.expect_column_values_to_be_between("age", min_value=0, max_value=150)

# Validate
results = df.validate()
print(results.success)
```

### Data Pipeline Monitoring

| Metric | Alert Threshold |
|--------|-----------------|
| Row count | ±20% from expected |
| Null rate | > 5% |
| Schema changes | Any unexpected |
| Freshness | > 2 hours stale |
| Duplicates | > 0.1% |
