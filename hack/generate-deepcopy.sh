set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=.
CODEGEN_PKG=./code-generator
THIS_PKG="crds-controller"

source "${CODEGEN_PKG}/kube_codegen.sh"

kube::codegen::gen_helpers \
    --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.go.txt" \
    "${SCRIPT_ROOT}/pkg/apis"