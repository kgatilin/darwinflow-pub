package app

import (
	"time"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// entityAdapter adapts pluginsdk.IExtensible to domain.IExtensible
// This allows plugins using the SDK to provide entities that work with internal app code
type entityAdapter struct {
	inner pluginsdk.IExtensible
}

// newEntityAdapter wraps an SDK entity to implement domain interfaces
func newEntityAdapter(sdkEntity pluginsdk.IExtensible) domain.IExtensible {
	return &entityAdapter{inner: sdkEntity}
}

func (e *entityAdapter) GetID() string {
	return e.inner.GetID()
}

func (e *entityAdapter) GetType() string {
	return e.inner.GetType()
}

func (e *entityAdapter) GetCapabilities() []string {
	return e.inner.GetCapabilities()
}

func (e *entityAdapter) GetField(name string) interface{} {
	return e.inner.GetField(name)
}

func (e *entityAdapter) GetAllFields() map[string]interface{} {
	return e.inner.GetAllFields()
}

// Check if entity also implements IHasContext
func (e *entityAdapter) GetContext() *domain.EntityContext {
	hasContext, ok := e.inner.(pluginsdk.IHasContext)
	if !ok {
		return nil
	}

	sdkCtx := hasContext.GetContext()
	if sdkCtx == nil {
		return nil
	}

	// Adapt SDK EntityContext to domain EntityContext
	return &domain.EntityContext{
		RelatedEntities: sdkCtx.RelatedEntities,
		LinkedFiles:     sdkCtx.LinkedFiles,
		RecentActivity:  adaptActivityRecords(sdkCtx.RecentActivity),
		Metadata:        sdkCtx.Metadata,
	}
}

// Check if entity also implements ITrackable
func (e *entityAdapter) GetStatus() string {
	trackable, ok := e.inner.(pluginsdk.ITrackable)
	if !ok {
		return ""
	}
	return trackable.GetStatus()
}

func (e *entityAdapter) GetProgress() float64 {
	trackable, ok := e.inner.(pluginsdk.ITrackable)
	if !ok {
		return 0
	}
	return trackable.GetProgress()
}

func (e *entityAdapter) IsBlocked() bool {
	trackable, ok := e.inner.(pluginsdk.ITrackable)
	if !ok {
		return false
	}
	return trackable.IsBlocked()
}

func (e *entityAdapter) GetBlockReason() string {
	trackable, ok := e.inner.(pluginsdk.ITrackable)
	if !ok {
		return ""
	}
	return trackable.GetBlockReason()
}

// Check if entity also implements ISchedulable (not yet in SDK)
func (e *entityAdapter) GetStartDate() *time.Time {
	// ISchedulable not yet implemented in SDK
	// Future: check if entity implements it
	return nil
}

func (e *entityAdapter) GetDueDate() *time.Time {
	// ISchedulable not yet implemented in SDK
	// Future: check if entity implements it
	return nil
}

func (e *entityAdapter) IsOverdue() bool {
	// ISchedulable not yet implemented in SDK
	// Future: check if entity implements it
	return false
}

// Check if entity also implements IRelatable (not yet in SDK)
func (e *entityAdapter) GetRelated(entityType string) []string {
	// IRelatable not yet implemented in SDK
	// Future: check if entity implements it
	return nil
}

func (e *entityAdapter) GetAllRelations() map[string][]string {
	// IRelatable not yet implemented in SDK
	// Future: check if entity implements it
	return nil
}

// adaptActivityRecords converts SDK activity records to domain activity records
func adaptActivityRecords(sdkRecords []pluginsdk.ActivityRecord) []domain.ActivityRecord {
	if sdkRecords == nil {
		return nil
	}

	domainRecords := make([]domain.ActivityRecord, len(sdkRecords))
	for i, sdkRec := range sdkRecords {
		// Map SDK fields to domain fields
		details := make(map[string]interface{})
		if sdkRec.Description != "" {
			details["description"] = sdkRec.Description
		}
		if sdkRec.Actor != "" {
			details["actor"] = sdkRec.Actor
		}

		domainRecords[i] = domain.ActivityRecord{
			Timestamp:   sdkRec.Timestamp,
			Type:        sdkRec.Type,
			Description: sdkRec.Description,
			Actor:       sdkRec.Actor,
		}
	}
	return domainRecords
}

// adaptEntities wraps SDK entities in domain adapters
func adaptEntities(sdkEntities []pluginsdk.IExtensible) []domain.IExtensible {
	if sdkEntities == nil {
		return nil
	}

	domainEntities := make([]domain.IExtensible, len(sdkEntities))
	for i, sdkEntity := range sdkEntities {
		domainEntities[i] = newEntityAdapter(sdkEntity)
	}
	return domainEntities
}

// adaptEntityQuery converts domain EntityQuery to SDK EntityQuery
func adaptEntityQuery(domainQuery domain.EntityQuery) pluginsdk.EntityQuery {
	return pluginsdk.EntityQuery{
		EntityType: domainQuery.EntityType,
		Filters:    domainQuery.Filters,
		Limit:      domainQuery.Limit,
		Offset:     domainQuery.Offset,
		SortBy:     domainQuery.SortBy,
		SortDesc:   domainQuery.SortDesc,
	}
}

// adaptEntityTypeInfo converts SDK EntityTypeInfo to domain EntityTypeInfo
func adaptEntityTypeInfo(sdkInfo pluginsdk.EntityTypeInfo) domain.EntityTypeInfo {
	return domain.EntityTypeInfo{
		Type:              sdkInfo.Type,
		DisplayName:       sdkInfo.DisplayName,
		DisplayNamePlural: sdkInfo.DisplayNamePlural,
		Capabilities:      sdkInfo.Capabilities,
		Icon:              sdkInfo.Icon,
	}
}

// adaptEntityTypeInfos converts slice of SDK EntityTypeInfo to domain EntityTypeInfo
func adaptEntityTypeInfos(sdkInfos []pluginsdk.EntityTypeInfo) []domain.EntityTypeInfo {
	if sdkInfos == nil {
		return nil
	}

	domainInfos := make([]domain.EntityTypeInfo, len(sdkInfos))
	for i, sdkInfo := range sdkInfos {
		domainInfos[i] = adaptEntityTypeInfo(sdkInfo)
	}
	return domainInfos
}
