Sync github and linux users
===========================

This will create linux users for each of your github team users, sync their 
github authorized_keys file, and disable them if they ever leave your 
organization. Set it on a timer and let it run regularly.

Configuration file:
```yaml
# githubsync.yaml
github:
  org-slug: <slug>
  team-slug: <slug>
  token: <token>
```

Caveats:
========
- If there are conflicts, github wins
- Users will be created with UID between 5000 and 6000
- Users will be added to the 'sudo' group
- Users will have their shell set to /bin/bash
- Users will be created without passwords
- Users will have their gecos field set to GitHub
