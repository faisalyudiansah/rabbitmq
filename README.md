# rabbitmq
background-job

```sql
CREATE TABLE jobs (
    id SERIAL PRIMARY KEY,
    type VARCHAR(50),
    payload JSONB,
    status VARCHAR(20) DEFAULT 'pending',
    retry INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP
);
```