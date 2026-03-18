package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"robbert/kube-portal/templates"
)

func urlScheme(ingress networkingv1.Ingress, host string) string {
	for _, tls := range ingress.Spec.TLS {
		if slices.Contains(tls.Hosts, host) {
			return "https"
		}
	}
	return "http"
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	kubeconfig := filepath.Join(homeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	ingressi, err := clientset.NetworkingV1().Ingresses("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	coolLinks := []templates.Link{}
	for _, ingress := range ingressi.Items {
		icon := ""
		label := ""

		for i, v := range ingress.Annotations {
			if strings.HasPrefix(i, "kube-portal/icon") {
				icon = v
				continue
			}

			if strings.HasPrefix(i, "kube-portal/label") {
				label = v
				continue
			}
		}

		for _, ingressRule := range ingress.Spec.Rules {
			scheme := urlScheme(ingress, ingressRule.Host)

			coolLinks = append(coolLinks, templates.Link{Label: label, Image: icon, Url: scheme + "://" + ingressRule.Host})
		}

	}

	viewModel := templates.DashboardViewModel{Links: coolLinks}
	templates.Home(viewModel).Render(r.Context(), w)
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("templates/layout"))))
	http.HandleFunc("/", viewHandler)

	http.ListenAndServe(":30000", nil)
}

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic("could not determine home directory")
	}
	return home
}
