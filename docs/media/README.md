# Documentation media

Every public project visual in this directory and `brand/source` is registered
in `manifest.json`. The manifest records whether an image shows current software
or a planned concept, its parsed dimensions, digest, size budget, public status,
source revision or generation record, post-processing, replacement history,
rights basis, reviewer role, and accessible description.

Rules:

- A current screenshot must be captured from the named repository revision and
  record the command or scenario shown.
- A planned concept must say `Planned` in the image, adjacent caption, or both.
- Generated and project-authored concepts cannot be presented as implemented UI.
- Generated concepts record the generator class, model disclosure available to
  the tool, sanitized prompt, source-asset digest, inputs, and post-processing.
- Brand files record their separate clearance state. Public asset approval is
  not trademark clearance.
- No file may contain a secret, username, machine name, absolute local path,
  account detail, private event reference, or tracking URL.
- Optimize losslessly where practical. A larger replacement needs a justified
  byte budget in the manifest.
- Update the digest and review fields whenever bytes change.
- The repository license applies only after the rights status permits public
  repository distribution.

Run the media check from the repository root:

```powershell
pwsh ./scripts/check-media.ps1
```

The check rejects missing files, digest drift, byte-budget overruns, unreferenced
README images, and orphan media.
