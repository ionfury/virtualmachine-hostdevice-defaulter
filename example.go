package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func handleMutate(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not read request body: %v", err), http.StatusBadRequest)
		return
	}

	// Parse the AdmissionReview request.
	var admissionReviewReq admissionv1.AdmissionReview
	if err := json.Unmarshal(body, &admissionReviewReq); err != nil {
		http.Error(w, fmt.Sprintf("could not unmarshal request: %v", err), http.StatusBadRequest)
		return
	}

	// Prepare the AdmissionReview response.
	admissionReviewResp := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
	}

	// Default response is to allow the Pod creation/modification.
	admissionReviewResp.Response = &admissionv1.AdmissionResponse{
		UID:     admissionReviewReq.Request.UID,
		Allowed: true,
	}

	// Decode the Pod object from the request.
	var pod corev1.Pod
	if err := json.Unmarshal(admissionReviewReq.Request.Object.Raw, &pod); err != nil {
		http.Error(w, fmt.Sprintf("could not unmarshal pod object: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if the annotation is already set, if not, mutate the Pod.
	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}
	if _, ok := pod.Annotations["mutated"]; !ok {
		patch := []map[string]interface{}{
			{
				"op":    "add",
				"path":  "/metadata/annotations/mutated",
				"value": "true",
			},
		}
		patchBytes, err := json.Marshal(patch)
		if err != nil {
			http.Error(w, fmt.Sprintf("could not marshal patch: %v", err), http.StatusInternalServerError)
			return
		}
		admissionReviewResp.Response.Patch = patchBytes
		admissionReviewResp.Response.PatchType = new(admissionv1.PatchType)
		*admissionReviewResp.Response.PatchType = admissionv1.PatchTypeJSONPatch
	}

	// Respond to the admission request.
	respBytes, err := json.Marshal(admissionReviewResp)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not marshal response: %v", err), http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(respBytes); err != nil {
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/mutate", handleMutate)
	fmt.Printf("Starting webhook server...\n")
	if err := http.ListenAndServeTLS(":443", "/path/to/tls.crt", "/path/to/tls.key", nil); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to listen and serve webhook server: %v\n", err)
		os.Exit(1)
	}
}
