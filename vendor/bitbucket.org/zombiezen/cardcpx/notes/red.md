# RED Camera Directory Structure

Given a path `A002_0908FT.RDM/A002_C001_0908R4.RDC/A002_C001_0908R4_003.R3D`:

* `A002` is the camera `A`, card `002`.
* `0908` is the date (September 8th)
* `C001` is clip/take #1
* `R4` and `FT` are arbitrary hex distinguishers to make them unique-ish
* `003` indicates that this is the 3rd chunk of the R3D clip.  Each clip can be
  \>4GB, so the RED camera splits them up into 2GB chunks.

## Example

```
/path/to/card/
  digital_magazine.bin
  digital_magdynamic.bin
  A002_0908FT.RDM/
    A002_C001_0908R4.RDC/
      A002_C001_0908R4_001.R3D
      A002_C001_0908R4_002.R3D
      A002_C001_0908R4_003.R3D
    A002_C002_09081S.RDC/
      A002_C002_09081S_001.R3D
      A002_C002_09081S_002.R3D
      A002_C002_09081S_003.R3D
      A002_C002_09081S_004.R3D
```

## Sources

1. http://wolfcrow.com/blog/how-to-work-with-redcode-raw-r3d-footage/
2. http://www.assimilateinc.com/pdfs/SCRATCH_and_RED_Workflow.pdf
3. http://www.davidelkins.com/cam/manuals/manual_files/red/epic_manual_3.3.pdf
