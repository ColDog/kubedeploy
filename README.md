# Kubedeploy

A simple dockerless deployment tool for building applications on top of Kubernetes.

Why did I build this?

## Getting Started

Prerequisite: A running Kubernetes cluster.

1. Install [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/).
2. Install [`helm`](https://docs.helm.sh/using_helm/).

Install the core package:

    version=$(curl https://raw.githubusercontent.com/ColDog/kubedeploy/master/version)
    helm install --namespace kubedeploy --name kubedeploy https://github.com/ColDog/kubedeploy/releases/download/${version}/kubedeploy-${version}.tgz

Install the CLI:

    curl https://raw.githubusercontent.com/ColDog/kubedeploy/master/install.sh | sudo bash

Create an `app.yaml` file:

```bash
cat > app.yaml <<EOF
name: example
source: index.js
runtime: node
EOF
```

Create an `index.js` file:

```bash
cat > index.js <<EOF
const http = require('http')
const port = 8080

const requestHandler = (req, res) => { res.end('Hello Node.js Server!') }

const server = http.createServer(requestHandler)
server.listen(port, (err) => {
  if (err) throw err;
  console.log(`server is listening on ${port}`)
})
EOF
```

Run `kubedeploy deploy` to deploy your application to Kubernetes!
