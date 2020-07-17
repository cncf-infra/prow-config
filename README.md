## High level overview

[Project board](https://github.com/cncf/apisnoop/projects/29)

The conformance gate is comprised of 2 external Prow plugins that validates conformance PR's to ensure they meet all criteria set out by [conformance/instructions](https://github.com/cncf/k8s-conformance/blob/master/instructions.md#uploading)

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
- Label release-X.Y OR needs-release + docs
- Check Folder name /vX.X/ matches version
- Use Folder /vX.X/${product} as the product and base folder
- Check that PRODUCT.yaml has required fields
- Check that e2e.log kube-apiserver version matches

#### verify-conformance-tests

Performs the following functionality:

- Test results in junit_01.xml must include all required tests for the specified release of Kubernetes
- e2e.log confirms that no tests failed.
- TODO: Test result logs in the e2e.log file need to show that all tests present in the junit results file are present in the e2e.log file,


Once both plugins pass the PR will be given a conformance lable and assigned to a human for final review
