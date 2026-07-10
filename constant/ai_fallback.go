package constant

const (
	FallbackReasonDisabled            = "ai_disabled"
	FallbackReasonNoDriverNote        = "no_driver_note"
	FallbackReasonEarlyNoteIgnored    = "early_note_ignored_to_save_cost"
	FallbackReasonTemplateSufficient  = "template_sufficient"
	FallbackReasonTimeout             = "ai_timeout"
	FallbackReasonConnectionError     = "ai_connection_error"
	FallbackReasonInvalidResponse     = "ai_invalid_response"
	FallbackReasonLowConfidence       = "ai_confidence_below_threshold"
	FallbackReasonStructuralRuleMatch = "structural_rule_match"
	FallbackReasonNoteNotActionable   = "note_not_actionable"
)
