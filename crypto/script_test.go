package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScript(t *testing.T) {
	priv := GeneratePrivateKey()
	msg := []byte("Helo world")

	sig := priv.Sign(msg)

	var scriptPub ScriptPubKey = priv.Public()
	scriptSig := ScriptSig{
		Sig:    sig,
		Signer: priv.Public(),
	}

	assert.True(t, Eval(scriptSig, scriptPub, msg))

	invalidPriv := GeneratePrivateKey()

	invalidScriptSig := ScriptSig{
		Sig:    invalidPriv.Sign(msg),
		Signer: invalidPriv.Public(),
	}
	assert.False(t, Eval(invalidScriptSig, scriptPub, msg))
}
