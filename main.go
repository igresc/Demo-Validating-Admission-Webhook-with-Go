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
	logger.Printf("received message on validate")

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
		msg := fmt.Sprintf("error getting admission review from request: %v", err)
		w.WriteHeader(400)
		w.Write([]byte(msg))
		logger.Fatal(msg)
	}

	logger.Println(admissionReviewRequest)

	// TODO: verify the resource is a pod
	// TODO: check resource limit is set to allow the creation

	// Generate the minimal response to allow pod creation
	admissionResponse := &admissionv1.AdmissionResponse{}
	admissionResponse.Allowed = true
	admissionResponse.UID = admissionReviewRequest.Request.UID

	var admissionReviewResponse admissionv1.AdmissionReview
	admissionReviewResponse.Response = admissionResponse
	admissionReviewResponse.SetGroupVersionKind(admissionReviewRequest.GroupVersionKind())

	resp, _ := json.Marshal(admissionReviewResponse)
	// if err != nil {
	// 	msg := fmt.Sprintf("error marshalling response json: %v", err)
	// 	logger.Printf(msg)
	// 	w.WriteHeader(500)
	// 	w.Write([]byte(msg))
	// 	return
	// }

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// Tie the command-line flag to the intervalFlag variable and
// set a usage message.
func init() {

	flag.StringVar(&tlsKey, "tls-key", "/etc/cert/tls.key", "path to the TLS private key (default: /etc/cert/tls.key)")
	flag.StringVar(&tlsCert, "tls-cert", "/etc/cert/tls.cert", "path to the TLS certificate (default: /etc/cert/tls.crt)")
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
	//logger.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))

}
