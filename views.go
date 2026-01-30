package pachca

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
)

const (
	ViewTypeModal = "modal"

	BlockHeader    = "header"
	BlockPlainText = "plain_text"
	BlockMarkdown  = "markdown"
	BlockDivider   = "divider"
	BlockInput     = "input"
	BlockSelect    = "select"
	BlockRadio     = "radio"
	BlockCheckbox  = "checkbox"
	BlockDate      = "date"
	BlockTime      = "time"
	BlockFileInput = "file_input"
)

type Views struct {
	client *resty.Client
}

type ViewRequest struct {
	Type            string `json:"type"` // modal
	TriggerID       string `json:"trigger_id"`
	PrivateMetadata string `json:"private_metadata,omitempty"`
	CallbackID      string `json:"callback_id,omitempty"`
	View            View   `json:"view"`
}

type View struct {
	Title      string      `json:"title"`
	CloseText  string      `json:"close_text,omitempty"`
	SubmitText string      `json:"submit_text,omitempty"`
	Blocks     []ViewBlock `json:"blocks"`
}

type ViewBlock struct {
	Type string `json:"type"`

	// common
	Text string `json:"text,omitempty"`

	// input / select / radio / checkbox / date / time / file_input
	Name     string `json:"name,omitempty"`
	Label    string `json:"label,omitempty"`
	Required bool   `json:"required,omitempty"`
	Hint     string `json:"hint,omitempty"`

	// input
	Placeholder  string `json:"placeholder,omitempty"`
	Multiline    bool   `json:"multiline,omitempty"`
	InitialValue string `json:"initial_value,omitempty"`
	MinLength    int    `json:"min_length,omitempty"`
	MaxLength    int    `json:"max_length,omitempty"`

	// select / radio / checkbox
	Options []ViewOption `json:"options,omitempty"`

	// date
	InitialDate string `json:"initial_date,omitempty"`

	// time
	InitialTime string `json:"initial_time,omitempty"`

	// file_input
	Filetypes []string `json:"filetypes,omitempty"`
	MaxFiles  int      `json:"max_files,omitempty"`
}

type ViewOption struct {
	Text        string `json:"text"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`

	// use ONLY ONE of these depending on block type
	Checked  bool `json:"checked,omitempty"`  // radio / checkbox
	Selected bool `json:"selected,omitempty"` // select
}

func (v *Views) Open(ctx context.Context, req *ViewRequest) error {
	if v == nil || v.client == nil {
		return errors.New("views client is not initialized")
	}

	if req.TriggerID == "" {
		return errors.New("trigger_id is required")
	}

	if req.Type == "" {
		req.Type = ViewTypeModal
	}

	var resp struct {
		OK    bool   `json:"ok"`
		Error string `json:"error,omitempty"`
	}

	r, err := v.client.
		R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&resp).
		Post("/views/open")
	if err != nil {
		return fmt.Errorf("views.open request failed: %w", err)
	}

	if r.IsError() {
		return fmt.Errorf(
			"views.open http error: status=%d body=%s",
			r.StatusCode(),
			string(r.Body()),
		)
	}

	if !resp.OK {
		return fmt.Errorf("views.open api error: %s", resp.Error)
	}

	return nil
}
