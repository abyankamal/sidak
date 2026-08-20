package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/abyankamal/sidak/backend/internal/repository"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

type SchemaCache struct {
	repo    *repository.TemplateRepository
	schemas sync.Map // map[string]*jsonschema.Schema
}

func NewSchemaCache(repo *repository.TemplateRepository) *SchemaCache {
	return &SchemaCache{
		repo: repo,
	}
}

func (sc *SchemaCache) LoadAll(ctx context.Context) error {
	templates, err := sc.repo.GetActiveTemplates(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch templates: %w", err)
	}

	for _, t := range templates {
		if err := sc.compileAndStore(t.LayananID, t.SkemaJSON); err != nil {
			return fmt.Errorf("failed to compile schema for %s: %w", t.LayananID, err)
		}
	}

	return nil
}

func (sc *SchemaCache) compileAndStore(layananID string, rawJSON json.RawMessage) error {
	compiler := jsonschema.NewCompiler()
	url := fmt.Sprintf("schema://%s.json", layananID)

	if err := compiler.AddResource(url, strings.NewReader(string(rawJSON))); err != nil {
		return err
	}

	schema, err := compiler.Compile(url)
	if err != nil {
		return err
	}

	sc.schemas.Store(layananID, schema)
	return nil
}

func (sc *SchemaCache) Validate(layananID string, rawData json.RawMessage) error {
	val, ok := sc.schemas.Load(layananID)
	if !ok {
		// Try fetching from DB on-demand if not cached
		tmpl, err := sc.repo.GetByID(context.Background(), layananID)
		if err != nil || tmpl == nil {
			return fmt.Errorf("layanan_id '%s' tidak ditemukan atau tidak aktif", layananID)
		}
		if err := sc.compileAndStore(tmpl.LayananID, tmpl.SkemaJSON); err != nil {
			return fmt.Errorf("gagal mengompilasi skema layanan: %w", err)
		}
		val, _ = sc.schemas.Load(layananID)
	}

	schema, ok := val.(*jsonschema.Schema)
	if !ok {
		return fmt.Errorf("skema invalid untuk layanan %s", layananID)
	}

	var data any
	if err := json.Unmarshal(rawData, &data); err != nil {
		return fmt.Errorf("data isian bukan format JSON valid: %w", err)
	}

	if err := schema.Validate(data); err != nil {
		return fmt.Errorf("validasi data isian gagal: %w", err)
	}

	return nil
}

func (sc *SchemaCache) Refresh(ctx context.Context) error {
	return sc.LoadAll(ctx)
}
