package function

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/sigstore/cosign/cmd/cosign/cli"
	"github.com/sigstore/cosign/cmd/cosign/cli/fulcio"
	"github.com/sigstore/cosign/pkg/cosign"
	"net/http"
	"os"
	"path/filepath"
)

type ImageVerificationReq struct {
	Image string
}

type ImageVerificationResp struct {
	Verified            bool   `json:"verified"`
	VerificationMessage string `json:"verification_message"`
}

func Handle(w http.ResponseWriter, r *http.Request) {
	var body ImageVerificationReq
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := context.TODO()

	wDir, err := os.Getwd()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	key, err := cli.LoadPublicKey(ctx, filepath.Join(wDir, "cosign.pub"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ref, err := name.ParseReference(body.Image)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	co := &cosign.CheckOpts{
		RootCerts:   fulcio.Roots,
		SigVerifier: key,
	}

	w.Header().Set("Content-Type", "application/json")
	var resp ImageVerificationResp
	if _, err = cosign.Verify(ctx, ref, co); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp = ImageVerificationResp{
			Verified:            false,
			VerificationMessage: err.Error(),
		}
	} else {
		w.WriteHeader(http.StatusOK)
		resp = ImageVerificationResp{
			Verified:            true,
			VerificationMessage: fmt.Sprintf("valid signatures found for an image: %s", body.Image),
		}
	}

	respAsByte, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(respAsByte)
}
