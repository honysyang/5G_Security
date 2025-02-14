package detector

var ProcedureCodeToStringMap = map[int]string{
	21:  "id-NGSetup",
	22:  "id-InitialUEMessage",
	23:  "id-UEContextReleaseRequest",
	24:  "id-UEContextReleaseCommand",
	25:  "id-UEContextReleaseComplete",
	26:  "id-Reset",
	27:  "id-NGReset",
	28:  "id-ErrorIndication",
	29:  "id-OverloadStart",
	30:  "id-OverloadStop",
	31:  "id-UEContextModificationRequest",
	32:  "id-UEContextModificationResponse",
	33:  "id-UEContextModificationFailure",
	34:  "id-RRCInactiveTransitionReport",
	35:  "id-HandoverRequired",
	36:  "id-HandoverCommand",
	37:  "id-HandoverPreparationFailure",
	38:  "id-HandoverRequest",
	39:  "id-HandoverResponse",
	40:  "id-HandoverFailure",
	41:  "id-UEContextSuspendRequest",
	42:  "id-UEContextSuspendResponse",
	43:  "id-UEContextResumeRequest",
	44:  "id-UEContextResumeResponse",
	45:  "id-UEContextResumeFailure",
	46:  "id-AMFConfigurationUpdate",
	47:  "id-AMFConfigurationUpdateAcknowledge",
	48:  "id-AMFConfigurationUpdateFailure",
	49:  "id-UEContextModificationIndication",
	50:  "id-UEContextModificationConfirm",
	255: "id-PrivateMessage",
	// 添加更多的映射项
}

var CriticalityToStringMap = map[int]string{
	0: "reject",
	1: "ignore",
	2: "notify",
	// 添加更多的映射项
}
