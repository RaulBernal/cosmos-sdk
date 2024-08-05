package client

import (
	"time"

	"cosmossdk.io/x/auth/signing"
	txsigning "cosmossdk.io/x/tx/signing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
)

type (
	// TxEncodingConfig defines an interface that contains transaction
	// encoders and decoders
	TxEncodingConfig interface {
		TxEncoder() sdk.TxEncoder
		TxDecoder() sdk.TxDecoder
		TxJSONEncoder() sdk.TxEncoder
		TxJSONDecoder() sdk.TxDecoder
		MarshalSignatureJSON([]signingtypes.SignatureV2) ([]byte, error)
		UnmarshalSignatureJSON([]byte) ([]signingtypes.SignatureV2, error)
	}

	// TxConfig defines an interface a client can utilize to generate an
	// application-defined concrete transaction type. The type returned must
	// implement TxBuilder.
	TxConfig interface {
		TxEncodingConfig

		NewTxBuilder() TxBuilder
		WrapTxBuilder(sdk.Tx) (TxBuilder, error)
		SignModeHandler() *txsigning.HandlerMap
		SigningContext() *txsigning.Context
	}

	// TxBuilder defines an interface which an application-defined concrete transaction
	// type must implement. Namely, it must be able to set messages, generate
	// signatures, and provide canonical bytes to sign over. The transaction must
	// also know how to encode itself.
	TxBuilder interface {
		GetTx() signing.Tx

		SetMsgs(msgs ...sdk.Msg) error
		SetSignatures(signatures ...signingtypes.SignatureV2) error
		SetMemo(memo string)
		SetFeeAmount(amount sdk.Coins)
		SetFeePayer(feePayer sdk.AccAddress)
		SetGasLimit(limit uint64)
		SetTimeoutHeight(height uint64)
		SetTimeoutTimestamp(timestamp time.Time)
		SetUnordered(v bool)
		SetFeeGranter(feeGranter sdk.AccAddress)
		AddAuxSignerData(tx.AuxSignerData) error
	}

	// ExtendedTxBuilder extends the TxBuilder interface,
	// which is used to set extension options to be included in a transaction.
	ExtendedTxBuilder interface {
		SetExtensionOptions(extOpts ...*codectypes.Any)
	}
)
