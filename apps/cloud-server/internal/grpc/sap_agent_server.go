package grpc

import (
	"context"
	"io"
	"time"

	"github.com/NasTecSol/nembus-core/grpc/sapagentpb"
	"github.com/NasTecSol/nembus-core/usecase"
)

type SapAgentServer struct {
	sapagentpb.UnimplementedSapAgentServiceServer
	uc *usecase.SapAgentUseCase
}

func NewSapAgentServer(uc *usecase.SapAgentUseCase) *SapAgentServer {
	return &SapAgentServer{uc: uc}
}

func (s *SapAgentServer) ResolveTenant(ctx context.Context, req *sapagentpb.TenantResolveRequest) (*sapagentpb.TenantResolveResponse, error) {
	res, err := s.uc.ResolveTenant(ctx, req.GetTenantSlug())
	if err != nil {
		return nil, err
	}
	return &sapagentpb.TenantResolveResponse{
		TenantSlug:     res.TenantSlug,
		TenantName:     res.TenantName,
		OrganizationId: res.OrganizationID,
		DefaultStoreId: res.DefaultStoreID,
	}, nil
}

func (s *SapAgentServer) StreamMigrationIngest(stream sapagentpb.SapAgentService_StreamMigrationIngestServer) error {
	ctx := stream.Context()
	var lastAck *sapagentpb.MigrationIngestAck

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			if lastAck == nil {
				return stream.SendAndClose(&sapagentpb.MigrationIngestAck{Success: true})
			}
			return stream.SendAndClose(lastAck)
		}
		if err != nil {
			return err
		}

		orgID := int(chunk.GetOrganizationId())
		if orgID <= 0 {
			orgID = 1
		}

		resp, err := s.uc.IngestMigrationChunk(ctx, orgID, chunk.GetPayloadGzip())
		ack := &sapagentpb.MigrationIngestAck{
			Success:       err == nil && resp != nil && resp.Success,
			BatchId:       chunk.GetBatchId(),
			Domain:        chunk.GetDomain(),
			RecordsStaged: int32(0),
		}
		if resp != nil {
			ack.RecordsStaged = int32(resp.RecordsStaged)
			ack.RecordsMerged = int32(resp.RecordsMerged)
			ack.RecordsFailed = int32(resp.RecordsFailed)
			ack.RecordsSkipped = int32(resp.RecordsSkipped)
			ack.ErrorMessage = resp.ErrorMessage
			ack.Errors = resp.Errors
			if !resp.Success {
				ack.Success = false
			}
		}
		if err != nil {
			ack.Success = false
			ack.ErrorMessage = err.Error()
		}
		lastAck = ack
	}
}

func (s *SapAgentServer) StreamMasterDataUpsert(stream sapagentpb.SapAgentService_StreamMasterDataUpsertServer) error {
	ctx := stream.Context()
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		res, applyErr := s.uc.ApplyMasterDataUpsert(ctx, msg.GetTenantSlug(), msg.GetOrganizationId(), msg.GetOperation(), msg.GetPayloadJson())
		ack := &sapagentpb.MasterDataUpsertAck{
			Success:   applyErr == nil && res != nil && res.Success,
			Operation: msg.GetOperation(),
		}
		if res != nil {
			ack.EntityId = res.EntityID
			ack.ResultJson = res.ResultJSON
			if res.ErrorText != "" {
				ack.ErrorMessage = res.ErrorText
				ack.Success = false
			}
		}
		if applyErr != nil {
			ack.Success = false
			ack.ErrorMessage = applyErr.Error()
		}
		if err := stream.Send(ack); err != nil {
			return err
		}
	}
}

func (s *SapAgentServer) StreamSapOutbox(req *sapagentpb.SapOutboxRequest, stream sapagentpb.SapAgentService_StreamSapOutboxServer) error {
	bundles, err := s.uc.StreamSapOutbox(stream.Context(), req.GetTenantSlug(), int(req.GetLimit()))
	if err != nil {
		return err
	}
	for _, b := range bundles {
		if err := stream.Send(&sapagentpb.SapOutboxEvent{
			EntityType:   b.EntityType,
			EntityId:     b.EntityID,
			PayloadJson:  b.HeaderJSON,
			LinesJson:    b.LinesJSON,
			PaymentsJson: b.PaymentsJSON,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *SapAgentServer) AckSapOutbox(ctx context.Context, req *sapagentpb.AckSapOutboxRequest) (*sapagentpb.AckSapOutboxResponse, error) {
	items := make([]usecase.SapOutboxAckItem, 0, len(req.GetItems()))
	for _, it := range req.GetItems() {
		items = append(items, usecase.SapOutboxAckItem{
			EntityType:  it.GetEntityType(),
			EntityID:    it.GetEntityId(),
			SapDocEntry: int(it.GetSapDocEntry()),
			SapDocNum:   int(it.GetSapDocNum()),
		})
	}
	n, err := s.uc.AckSapOutbox(ctx, req.GetTenantSlug(), items)
	if err != nil {
		return nil, err
	}
	return &sapagentpb.AckSapOutboxResponse{Acknowledged: int32(n)}, nil
}

func (s *SapAgentServer) HoldSapOutbox(ctx context.Context, req *sapagentpb.HoldSapOutboxRequest) (*sapagentpb.HoldSapOutboxResponse, error) {
	err := s.uc.HoldSapOutbox(ctx, req.GetTenantSlug(), usecase.HoldSapOutboxInput{
		EntityType:        req.GetEntityType(),
		EntityID:          req.GetEntityId(),
		TransactionNumber: req.GetTransactionNumber(),
		StoreID:           req.GetStoreId(),
		WarehouseCode:     req.GetWarehouseCode(),
		TotalAmount:       req.GetTotalAmount(),
		Summary:           req.GetSummary(),
		ShortageItemsJSON: req.GetShortageItemsJson(),
		OrganizationID:    req.GetOrganizationId(),
	})
	if err != nil {
		return &sapagentpb.HoldSapOutboxResponse{Success: false}, err
	}
	return &sapagentpb.HoldSapOutboxResponse{Success: true}, nil
}

func (s *SapAgentServer) GetReconciliationSummary(ctx context.Context, req *sapagentpb.ReconciliationRequest) (*sapagentpb.ReconciliationSummary, error) {
	metrics, err := s.uc.GetReconciliationMetrics(ctx, req.GetTenantSlug(), int(req.GetOrganizationId()))
	if err != nil {
		return nil, err
	}

	out := &sapagentpb.ReconciliationSummary{
		RunId:             req.GetRunId(),
		OrganizationId:    req.GetOrganizationId(),
		GeneratedAtRfc3339: time.Now().Format(time.RFC3339),
	}
	for _, d := range metrics.Domains {
		out.Domains = append(out.Domains, &sapagentpb.DomainReconciliationRow{
			Domain:           string(d.Domain),
			SapSourceCount:   d.SAPSourceCount,
			TargetCount:      d.TargetCount,
			Difference:       d.Difference,
			SapNumericSum:    d.SAPNumericSum,
			TargetNumericSum: d.TargetNumericSum,
			SumDifference:    d.SumDifference,
			Status:           d.Status,
			Notes:            d.Notes,
		})
	}
	return out, nil
}
