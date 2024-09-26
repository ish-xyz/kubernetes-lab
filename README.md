# kubernetes-lab


**TODO:**

# bm coding
- add logging
- add some codebase testing (30%/40%)
- add config validation
    - fix missing namespace in chart installation error
- test edge cases:
    - misconfigurations
        - invalid helm charts (preMigration)

- revert if migration step failed
- reload systemd daemons at end of the migration?

# terraform/workflow
- upload binary on github.com
- create systemd unit for bootstrap manager
- download bootstrap manager via cloud init
- reprovision cluster via terraform and try workflow
- fix kubeconfig file permissions (640)
    - create kubernetes group
    - add user zero to kubernetes group


