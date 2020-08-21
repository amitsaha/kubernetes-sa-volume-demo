package main

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"io"
	"net/http"

	authv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func verifyToken(clientId string) (bool, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	ctx := context.TODO()
	tr := authv1.TokenReview{
		Spec: authv1.TokenReviewSpec{
			Token: clientId,
		},
	}
	result, err := clientset.AuthenticationV1().TokenReviews().Create(ctx, &tr, metav1.CreateOptions{})
	if err != nil {
		return false, err
	}

	// For educational purposes only, if you need to log, use logging package or similar
	spew.Dump(result.Status)

	if result.Status.Authenticated {
		return true, nil
	}
	return false, nil

}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	// Read the value of the client identifier from the request header
	clientId := r.Header.Get("X-Client-Id")
	if len(clientId) == 0 {
		http.Error(w, "X-Client-Id not supplied", http.StatusUnauthorized)
		return
	}
	authenticated, err := verifyToken(clientId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if !authenticated {
		http.Error(w, "Invalid token", http.StatusForbidden)
		return
	}
	io.WriteString(w, "Hello from service2. You have been authenticated")
}

func main() {

	http.HandleFunc("/", handleIndex)
	http.ListenAndServe(":8081", nil)
}
