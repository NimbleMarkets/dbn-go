// Copyright (c) 2024 Neomantra Corp

package dbn

type Visitor interface {
	OnMbp0(record *Mbp0Msg) error
	OnMbp1(record *Mbp1Msg) error
	OnMbp10(record *Mbp10Msg) error
	OnMbo(record *MboMsg) error

	OnOhlcv(record *OhlcvMsg) error
	OnCmbp1(record *Cmbp1Msg) error
	OnBbo(record *BboMsg) error

	OnImbalance(record *ImbalanceMsg) error
	OnStatMsg(record *StatMsg) error
	OnStatusMsg(record *StatusMsg) error
	OnInstrumentDefMsg(record *InstrumentDefMsg) error

	OnErrorMsg(record *ErrorMsg) error
	OnSystemMsg(record *SystemMsg) error
	OnSymbolMappingMsg(record *SymbolMappingMsg) error

	OnStreamEnd() error
}
