## High level overview

The conformance gate is comprised of 2 external Prow plugins that validates conformance PR's to ensure they meet all criteria set out by [conformance/instructions](https://github.com/cncf/k8s-conformance/blob/master/instructions.md#uploading)

[Project board](https://github.com/cncf/apisnoop/projects/29)

The plugins used in this gate:
1. verify-conformance-release
2. verify-conformance-tests

These plugins will currently run as pods on Prow, from there they will watch for any new PR's against [cncf-infra/k8s-conformance](https://github.com/cncf-infra/k8s-conformance/)

If both plugins pass the issue will be labled and assigned to a human for final review.

### Plugins explained

#### verify-conformance-release
Performs the following functionality:
- Run against each PR submitted to k8s-conformance
- Check PR Title for vX.X as the version
- Add label release-X.Y OR needs-release + docs
- Check Folder name /vX.X/ matches version
- Check that PRODUCT.yaml has required fields
- Check that e2e.log kube-apiserver version matches
- If the folder, product.yaml and e2e log pass we add a release-documents-checked label
- If any checks fail a not-verifiable label with clarifying comment gets added

#### verify-conformance-tests
- Run against each PR submitted to k8s-conformance
- Check that Test results in junit_01.xml include all required tests for the specified release of Kubernetes
- Check e2e.log to confirm that there are no failed tests
- If all tests from conformance.yaml is present in junit add tests-verified-v1.xx label
- If any tests are missing add a required-tests-missing label
- If there are no failed tests in e2e add a no-failed-tests-v1.x.x label
- If there is a failed test add the evidence-missing label
