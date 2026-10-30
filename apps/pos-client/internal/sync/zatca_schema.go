package sync

// Installed idempotently after Goose, including on databases that have already
// applied the old POS-extension migration. Pull/recovery writes suppress echo.
const ZatcaOutboxDDL = `
CREATE OR REPLACE FUNCTION trg_enqueue_zatca_sync() RETURNS trigger AS $$
DECLARE event_payload jsonb;
BEGIN
 IF current_setting('nembus.sync_origin',true)='cloud' THEN RETURN NEW; END IF;
 IF TG_TABLE_NAME='zatca_device_configs' THEN
  IF NEW.device_type<>'pos' THEN RETURN NEW; END IF;
  event_payload=jsonb_build_object('id',NEW.id,'device_serial',NEW.device_serial,'updated_at',NEW.updated_at);
 ELSE
  IF NOT EXISTS(SELECT 1 FROM zatca_device_configs WHERE id=NEW.device_config_id AND device_type='pos') THEN RETURN NEW; END IF;
  event_payload=jsonb_build_object('id',NEW.id,'device_config_id',NEW.device_config_id,'icv',NEW.icv,'updated_at',NEW.updated_at);
 END IF;
 INSERT INTO sync_queue(entity_type,entity_id,action,payload,status,priority)
 VALUES(TG_TABLE_NAME,NEW.id,TG_OP,event_payload,'pending',20);
 RETURN NEW;
END; $$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS trg_sync_zatca_device_configs ON zatca_device_configs;
CREATE TRIGGER trg_sync_zatca_device_configs AFTER INSERT OR UPDATE ON zatca_device_configs FOR EACH ROW EXECUTE FUNCTION trg_enqueue_zatca_sync();
DROP TRIGGER IF EXISTS trg_sync_zatca_document_chain ON zatca_document_chain;
CREATE TRIGGER trg_sync_zatca_document_chain AFTER INSERT OR UPDATE ON zatca_document_chain FOR EACH ROW EXECUTE FUNCTION trg_enqueue_zatca_sync();
CREATE INDEX IF NOT EXISTS idx_sync_queue_entity_identity ON sync_queue(entity_type,entity_id);
`

const backfillZatcaOutboxSQL = `
INSERT INTO sync_queue(entity_type,entity_id,action,payload,status,priority)
SELECT 'zatca_device_configs',d.id,'UPDATE',jsonb_build_object('id',d.id,'device_serial',d.device_serial,'updated_at',d.updated_at),'pending',20 FROM zatca_device_configs d
WHERE d.device_type='pos' AND EXISTS(SELECT 1 FROM local_device_config l WHERE l.pos_terminal_id=d.pos_terminal_id AND l.store_id=d.store_id)
AND NOT EXISTS(SELECT 1 FROM sync_queue q WHERE q.entity_type='zatca_device_configs' AND q.entity_id=d.id) ORDER BY d.id;
INSERT INTO sync_queue(entity_type,entity_id,action,payload,status,priority)
SELECT 'zatca_document_chain',c.id,'UPDATE',jsonb_build_object('id',c.id,'device_config_id',c.device_config_id,'icv',c.icv,'updated_at',c.updated_at),'pending',20 FROM zatca_document_chain c JOIN zatca_device_configs d ON d.id=c.device_config_id
WHERE d.device_type='pos' AND EXISTS(SELECT 1 FROM local_device_config l WHERE l.pos_terminal_id=d.pos_terminal_id AND l.store_id=d.store_id)
AND NOT EXISTS(SELECT 1 FROM sync_queue q WHERE q.entity_type='zatca_document_chain' AND q.entity_id=c.id) ORDER BY c.device_config_id,c.icv;
`
