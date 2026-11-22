# Warranty and documents

**What it does:** Keeps receipts and warranty papers with the product, shows how long each warranty runs, and stores the photo or PDF on your Waylight server.

## Using it

Add the *Warranty and documents* widget and enter a product with **+ New document**:

- **Product**, the **purchase day** and the **end of the warranty**. The last two are optional.
- A **note**, for example where the original receipt is.
- A **file**: a photo (JPEG, PNG or WebP) or a PDF of up to 10 MB. Choose a
  new file to replace the old one.

Each entry shows its warranty as running, ends soon (within 30 days) or
ended. Warranties that are still running come first, the soonest end on top,
then entries without a warranty, then the ended ones. The compact view names
the warranty that ends first and asks for attention when some end within 30
days. **Download** saves the file to your device; **✕** removes the entry
together with its file.

## Where the files are

Files are stored in Waylight's database in the data volume, like the drop
zone and the photo frame. Every backup and restore of the instance includes
them, and nothing is sent anywhere. The type of a file is checked by its
content, not by its name, so only photos and PDFs are accepted.

## Who sees what

New entries are private for members and admins, like notes, and can be shared
with a group or everyone. Only the person who entered a document can change
or remove it. Everyone who sees an entry can download its file.

## Configuration

No environment variables. `MODULES_DISABLED=documents` switches the module
off.

## Outbound requests

None.

## API

`/api/documents` (list, create, update, delete), `/api/documents/{id}/file`
(upload with `POST`, download with `GET`) and `/api/documents/compact` – see
`api/modules/documents.yaml`.
