package crypto

type ScriptSig struct {
	Sig    *Signature
	Signer *PublicKey
}

type ScriptPubKey *PublicKey

func Eval(scriptSig ScriptSig, scriptPub ScriptPubKey, msg []byte) bool {
	return scriptSig.Sig.Verify(scriptPub, msg)
}

func (sig ScriptSig) Use(pubKey *PublicKey) bool {
	return sig.Signer.Key.Equal(pubKey.Key)
}
