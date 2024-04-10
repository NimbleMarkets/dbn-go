// Copyright (c) 2024 Neomantra Corp

package dbn

type Visitor interface {
	OnMbp0(record *Mbp0) error
	OnOhlcv(record *Ohlcv) error
	OnImbalance(record *Imbalance) error
	OnStreamEnd() error
}
