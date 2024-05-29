// Copyright (c) 2024 Neomantra Corp

package dbn

type Visitor interface {
	OnMbp0(record *Mbp0) error
	OnMbp1(record *Mbp1Msg) error
	OnMbp10(record *Mbp10Msg) error
	OnMbo(record *MboMsg) error

	OnOhlcv(record *Ohlcv) error
	OnImbalance(record *Imbalance) error
	OnStatMsg(record *StatMsg) error

	OnErrorMsg(record *ErrorMsg) error
	OnSystemMsg(record *SystemMsg) error
	OnSymbolMappingMsg(record *SymbolMappingMsg) error

	OnStreamEnd() error
}
