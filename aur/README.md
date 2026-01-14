# AUR Package for Workspace CLI

This directory contains the PKGBUILD for the Arch User Repository (AUR).

## Initial Setup

1. Create an AUR account at https://aur.archlinux.org
2. Add your SSH public key to your AUR account
3. Clone the AUR repository:
   ```bash
   git clone ssh://aur@aur.archlinux.org/workspace-bin.git
   ```

## Updating the Package

After each release:

1. Update `pkgver` in PKGBUILD to match the new version
2. Generate checksums (or use SKIP for GitHub releases):
   ```bash
   updpkgsums
   ```
3. Generate .SRCINFO:
   ```bash
   makepkg --printsrcinfo > .SRCINFO
   ```
4. Test the package builds:
   ```bash
   makepkg -si
   ```
5. Push to AUR:
   ```bash
   cd workspace-bin
   cp ../PKGBUILD ../README.md .
   makepkg --printsrcinfo > .SRCINFO
   git add -A
   git commit -m "Update to v${pkgver}"
   git push
   ```

## Automation (Optional)

You can automate AUR updates using GitHub Actions with the
[aur-publish](https://github.com/marketplace/actions/aur-publish) action.

Required secrets:
- `AUR_SSH_PRIVATE_KEY`: Your AUR SSH private key
