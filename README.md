# kubernetes-lab


**TODO:**


# bm coding
- add logging
- parametrise initial wait time
- add some codebase testing (30%/40%)
- add initial checks
- add config validation
    - fix missing namespace in chart installation error
- add checks to pre and post migration workflows

# terraform/workflow
- upload binary on github.com
- create systemd unit for bootstrap manager
- download bootstrap manager via cloud init
- reprovision cluster via terraform and try workflow
- fix kubeconfig file permissions (640)
    - create kubernetes group
    - add user zero to kubernetes group
