-- Create schema and set search path
CREATE SCHEMA IF NOT EXISTS schedule;
SET search_path TO schedule;

-- Resource PV and SKU per node
CREATE TABLE schedule.agent_node_resource (
     id BIGSERIAL PRIMARY KEY,
     status INTEGER NOT NULL,
     resource_type INTEGER NOT NULL,
     location VARCHAR(255) NOT NULL,
     region VARCHAR(255) NOT NULL,
     gpu_type VARCHAR(255),
     driver_version VARCHAR(255),
     random_type VARCHAR(255) NULL,
     storage_type VARCHAR(255) NULL,
     cloud_type Integer NOT NULL,
     max_gpu_count INTEGER DEFAULT 0,
     max_vcpu INTEGER DEFAULT 0,
     max_vram INTEGER DEFAULT 0,
     max_ram INTEGER DEFAULT 0,
     max_persistent_volume INTEGER NOT NULL,
     network_upload NUMERIC(38,18) DEFAULT 0.0,
     network_download NUMERIC(38,18) DEFAULT 0.0,
     disk_read_speed NUMERIC(38,18) DEFAULT 0.0,
     disk_write_speed NUMERIC(38,18) DEFAULT 0.0,
     resource_md5 TEXT NOT NULL UNIQUE,
     created_at BIGINT NOT NULL,
     updated_at BIGINT,
     version INTEGER NOT NULL
);

COMMENT ON TABLE schedule.agent_node_resource IS 'Node resource specifications (GPU/CPU)';
COMMENT ON COLUMN schedule.agent_node_resource.cloud_type IS 'cloudType： 0-Secure, 1-Community';
COMMENT ON COLUMN schedule.agent_node_resource.status IS 'Status: 0-unavailable, 1-available';
COMMENT ON COLUMN schedule.agent_node_resource.resource_type IS 'Resource type: 1-GPU, 2-CPU';
COMMENT ON COLUMN schedule.agent_node_resource.gpu_type IS 'GPU model (e.g., NVIDIA A100)';
COMMENT ON COLUMN schedule.agent_node_resource.max_gpu_count IS 'Maximum GPU count per node';
COMMENT ON COLUMN schedule.agent_node_resource.max_vcpu IS 'Maximum virtual CPU cores per node';
COMMENT ON COLUMN schedule.agent_node_resource.max_vram IS 'Maximum video memory (VRAM) capacity in GB';
COMMENT ON COLUMN schedule.agent_node_resource.max_ram IS 'Maximum RAM capacity in GB';
COMMENT ON COLUMN schedule.agent_node_resource.max_persistent_volume IS 'Persistent storage capacity in GB';
COMMENT ON COLUMN schedule.agent_node_resource.resource_md5 IS 'Unique resource identifier MD5(resource_type:gpu_type:max_gpu_count:...)';
COMMENT ON COLUMN schedule.agent_node_resource.version IS 'Optimistic lock version number';

-- Node information table
CREATE TABLE schedule.agent_node_info (
     id BIGSERIAL PRIMARY KEY,
     agent_id VARCHAR(255) NOT NULL,
     agent_node_id VARCHAR(255) NOT NULL,
     node_status INTEGER NOT NULL,
     status INTEGER NOT NULL,
     resource_type INTEGER NOT NULL,
     metric_info JSONB NOT NULL,
     resource_free_info JSONB NOT NULL,
     base_image VARCHAR(255),
     location VARCHAR(255) NOT NULL,
     region VARCHAR(255) NOT NULL,
     gpu_type VARCHAR(255) NOT NULL,
     total_gpu_count INTEGER DEFAULT 0,
     total_vram INTEGER DEFAULT 0,
     total_ram INTEGER DEFAULT 0,
     total_vcpu INTEGER DEFAULT 0,
     node_resource_id BIGINT NOT NULL,
     created_at BIGINT NOT NULL,
     updated_at BIGINT,
     last_heart_beat BIGINT,
     version INTEGER NOT NULL
);

COMMENT ON TABLE schedule.agent_node_info IS 'Node information table';
COMMENT ON COLUMN schedule.agent_node_info.agent_id IS 'Kubernetes cluster ID';
COMMENT ON COLUMN schedule.agent_node_info.agent_node_id IS 'Kubernetes node ID';
COMMENT ON COLUMN schedule.agent_node_info.node_status IS 'Node status in Kubernetes (mapped from K8s status codes)';
COMMENT ON COLUMN schedule.agent_node_info.status IS 'Business status: -1-offline, 0-available, 1-busy';
COMMENT ON COLUMN schedule.agent_node_info.resource_type IS 'Resource type: 1-GPU, 2-CPU';
COMMENT ON COLUMN schedule.agent_node_info.base_image IS 'Base image version (e.g., Ubuntu 22.04)';
COMMENT ON COLUMN schedule.agent_node_info.total_gpu_count IS 'Total GPU count on node';
COMMENT ON COLUMN schedule.agent_node_info.total_vram IS 'Total video memory (VRAM) capacity in GB';
COMMENT ON COLUMN schedule.agent_node_info.total_ram IS 'Total RAM capacity in GB';
COMMENT ON COLUMN schedule.agent_node_info.total_vcpu IS 'Total virtual CPU cores';
COMMENT ON COLUMN schedule.agent_node_info.version IS 'Optimistic lock version number';

-- Foreign key constraint
ALTER TABLE schedule.agent_node_info
    ADD CONSTRAINT fk_agent_node_resource
 FOREIGN KEY (node_resource_id)
     REFERENCES schedule.agent_node_resource(id);

-- Agent information table
CREATE TABLE schedule.agent_info (
   id BIGSERIAL PRIMARY KEY,
   agent_id VARCHAR(255) NOT NULL UNIQUE,
   agent_status INTEGER NOT NULL,
   status INTEGER NOT NULL,
   cloud_type Integer NOT NULL,
   metric_info JSONB NOT NULL,
   resource_free_info JSONB NOT NULL,
   base_image VARCHAR(255),
   location VARCHAR(255) NOT NULL,
   region VARCHAR(255) NOT NULL,
   created_at BIGINT NOT NULL,
   updated_at BIGINT,
   last_heart_beat BIGINT,
   version INTEGER NOT NULL
);

COMMENT ON TABLE schedule.agent_info IS 'Agent information table';
COMMENT ON COLUMN schedule.agent_info.agent_status IS 'Agent status in Kubernetes';
COMMENT ON COLUMN schedule.agent_info.status IS 'Business status: -1-offline, 0-available, 1-busy';
COMMENT ON COLUMN schedule.agent_info.cloud_type IS 'cloudType： 0-Secure, 1-Community';
COMMENT ON COLUMN schedule.agent_info.base_image IS 'Base image version';
COMMENT ON COLUMN schedule.agent_info.version IS 'Optimistic lock version number';

-- Indexes for agent_info
CREATE INDEX idx_agent_info_metric ON schedule.agent_info USING gin (metric_info);
CREATE INDEX idx_agent_info_resource ON schedule.agent_info USING gin (resource_free_info);

-- Service information table
CREATE TABLE schedule.agent_service_info (
    id BIGSERIAL PRIMARY KEY,
    saas_pod_id BIGINT NOT NULL,
    agent_id VARCHAR(255),
    service_name VARCHAR(255),
    service_info JSONB,
    metric_info JSONB,
    status INTEGER NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT,
    service_updated_at BIGINT,
    last_heart_beat BIGINT,
    version INTEGER NOT NULL
);

COMMENT ON TABLE schedule.agent_service_info IS 'Service information table';
COMMENT ON COLUMN schedule.agent_service_info.saas_pod_id IS 'SaaS Pod ID (maps to K8s Service name)';
COMMENT ON COLUMN schedule.agent_service_info.service_info IS 'K8s service configuration (image, ports, resource limits, etc.)';
COMMENT ON COLUMN schedule.agent_service_info.status IS 'Service status: -1-failed, 0-deploying, 20-running';

-- Indexes for agent_service_info
CREATE UNIQUE INDEX udx_saas_pod_agent ON schedule.agent_service_info (saas_pod_id, agent_id);
CREATE INDEX idx_agent_service_agent ON schedule.agent_service_info (agent_id);

-- Service operation log table
CREATE TABLE schedule.service_operate_log (
     id BIGSERIAL PRIMARY KEY,
     service_info_id BIGINT NOT NULL,
     saas_pod_id BIGINT NOT NULL,
     schedule_info JSONB NOT NULL,
     operate_type INTEGER NOT NULL,
     status INTEGER NOT NULL,
     created_at BIGINT NOT NULL,
     updated_at BIGINT,
     completed_at BIGINT,
     version INTEGER NOT NULL
);

COMMENT ON TABLE schedule.service_operate_log IS 'Service operation logs';
COMMENT ON COLUMN schedule.service_operate_log.operate_type IS 'Operation type: 0-deploy, 1-update, 2-delete';
COMMENT ON COLUMN schedule.service_operate_log.status IS 'Operation status: 0-init, 1-completed';

-- Indexes for service_operate_log
CREATE INDEX idx_operate_log_service ON schedule.service_operate_log (service_info_id);
CREATE INDEX idx_operate_log_pod ON schedule.service_operate_log (saas_pod_id);

-- Routing rules table
CREATE TABLE schedule.route_rule (
   id BIGSERIAL PRIMARY KEY,
   rule_code VARCHAR(100) UNIQUE,
   description VARCHAR(255),
   enabled BOOLEAN DEFAULT TRUE,
   priority INTEGER DEFAULT 0,
   created_at BIGINT NOT NULL,
   updated_at BIGINT,
   version INTEGER NOT NULL
);

COMMENT ON TABLE schedule.route_rule IS 'Routing rules configuration';
COMMENT ON COLUMN schedule.route_rule.rule_code IS 'Rule code (e.g., config_match)';

-- Routing chain table
CREATE TABLE schedule.route_chain (
      id BIGSERIAL PRIMARY KEY,
      chain_name VARCHAR(100) UNIQUE,
      resource_type INTEGER NOT NULL,
      rules JSONB,
      enabled BOOLEAN DEFAULT TRUE,
      created_at BIGINT NOT NULL,
      updated_at BIGINT,
      version INTEGER NOT NULL
);

COMMENT ON TABLE schedule.route_chain IS 'Routing chain configuration';
COMMENT ON COLUMN schedule.route_chain.resource_type IS 'Resource type: 1-GPU, 2-CPU';
COMMENT ON COLUMN schedule.route_chain.rules IS 'Rule execution order (e.g., ["config_match", "resource_score"])';