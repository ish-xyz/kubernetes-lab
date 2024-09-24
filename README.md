# kubernetes-lab


**TODO:**


# bm coding
- add logging
- apply resources only in leader if says so (also add option for it)
- parametrise initial wait time
- parametries interval/retries for http and kubectl checks
- add some codebase testing (30%/40%)
- add initial checks
- add final checks functionality
- add config validation
    - fix missing namespace in chart installation error

# terraform/workflow
- upload binary on github.com
- create systemd unit for bootstrap manager
- download bootstrap manager via cloud init
- reprovision cluster via terraform and try workflow
- fix kubeconfig file permissions (640)
    - create kubernetes group
    - add user zero to kubernetes group
