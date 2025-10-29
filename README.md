# rabbitmq
background-job

```sql
CREATE TABLE jobs (
    id SERIAL PRIMARY KEY,
    type TEXT NOT NULL,
    payload JSONB NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending','success','failed')),
    retry_count INT DEFAULT 0,
    max_retry INT DEFAULT 3,
    last_error TEXT DEFAULT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```