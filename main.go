package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	tlsCert       string
	tlsKey        string
	port          int
	logger        = log.New(os.Stdout, "http: ", log.LstdFlags)
	runtimeScheme = runtime.NewScheme()
	codecFactory  = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecFactory.UniversalDeserializer()
)

func serveValidate(w http.ResponseWriter, r *http.Request) {
	logger.Printf("Received message on validate")

	// Recieve http request and check is not empty
	var body []byte
	if r.Body != nil {
		requestData, err := io.ReadAll(r.Body)
		if err != nil {
			return
		}
		body = requestData
	}

	// Parse the http request as AdmissionReview
	admissionReviewRequest := &admissionv1.AdmissionReview{}
	_, _, err := deserializer.Decode(body, nil, admissionReviewRequest)
	if err != nil {
		msg := fmt.Sprintf("Error getting admission review from request: %v", err)
		logger.Println(msg)
		w.WriteHeader(400)
		w.Write([]byte(msg))
		return
	}

	// TODO: verify the resource is a pod

	// Decode pod from the AdmissionReview
	raw := admissionReviewRequest.Request.Object.Raw
	pod := corev1.Pod{}
	_, _, err = deserializer.Decode(raw, nil, &pod)
	if err != nil {
		msg := fmt.Sprintf("Error decoding pod object: %v", err)
		logger.Println(msg)
		w.WriteHeader(500)
		w.Write([]byte(msg))
		return
	}

	logger.Println("Pod:", pod.Name)

	// Generate the response to allow pod creation
	// if the resource limit is set
	admissionResponse := &admissionv1.AdmissionResponse{}
	admissionResponse.Allowed = true

	// TODO: check all containers in the pod not only the first
	if pod.Spec.Containers[0].Resources.Limits.Memory().Value() <= 0 {
		logger.Println("pod nod not have memory resorce limits")
		// Memory() returns 0 if unspecified
		admissionResponse.Allowed = false
		admissionResponse.Result = &metav1.Status{Message: "Missing resource memory limit"}
	} else {
		logger.Println("Pod allowed since has memory limits")
	}
	logger.Println(pod.Spec.Containers[0].Resources.Limits)

	// Generate the admissionReview used for the reponse
	var admissionReviewResponse admissionv1.AdmissionReview
	admissionResponse.UID = admissionReviewRequest.Request.UID
	admissionReviewResponse.Response = admissionResponse
	admissionReviewResponse.SetGroupVersionKind(admissionReviewRequest.GroupVersionKind())

	resp, err := json.Marshal(admissionReviewResponse)
	if err != nil {
		msg := fmt.Sprintf("Error marshalling response json: %v", err)
		logger.Printf(msg)
		w.WriteHeader(500)
		w.Write([]byte(msg))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// Tie the command-line flags to the corresponding variables (tlsKey, tlsCert, port)
func init() {
	flag.StringVar(&tlsKey, "tls-key", "/etc/certs/tls.key", "path to the TLS private key (default: /etc/certs/tls.key)")
	flag.StringVar(&tlsCert, "tls-cert", "/etc/certs/tls.crt", "path to the TLS certificate (default: /etc/certs/tls.crt)")
	flag.IntVar(&port, "port", 443, "Port for the webhook server")
	flag.Parse()
}

// Main entrypoint to the server.
// Creates the HTTPS server with the tls certs.
// Creates a route to /validate which is the path used by k8s validating webhooks
func main() {

	http.HandleFunc("/validate", serveValidate)
	http.HandleFunc("/readyz", func(w http.ResponseWriter, req *http.Request) { w.Write([]byte("ok")) })
	logger.Printf("Server started on port %d ...\n", port)
	logger.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%d", port), tlsCert, tlsKey, nil))

}
