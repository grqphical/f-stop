const IMAGE_MIME_TYPES: Record<string, string[]> = {
    // Common web formats
    'image/apng': ['apng'],
    'image/avif': ['avif'],
    'image/bmp': ['bmp'],
    'image/gif': ['gif'],
    'image/jpeg': ['jpg', 'jpeg', 'jpe'],
    'image/png': ['png'],
    'image/svg+xml': ['svg', 'svgz'],
    'image/webp': ['webp'],
    'image/x-icon': ['ico'],
    'image/vnd.microsoft.icon': ['ico'],

    // Modern / high-efficiency formats
    'image/heic': ['heic'],
    'image/heif': ['heif'],
    'image/jxl': ['jxl'],
    'image/jp2': ['jp2'],
    'image/jpx': ['jpx'],
    'image/jpm': ['jpm'],
    'image/ktx': ['ktx'],
    'image/ktx2': ['ktx2'],

    // Print / professional
    'image/tiff': ['tif', 'tiff'],
    'image/vnd.adobe.photoshop': ['psd'],
    'image/x-adobe-dng': ['dng'],

    // Camera RAW (non-standard but common)
    'image/x-canon-cr2': ['cr2'],
    'image/x-canon-crw': ['crw'],
    'image/x-nikon-nef': ['nef'],
    'image/x-sony-arw': ['arw'],
    'image/x-olympus-orf': ['orf'],
    'image/x-fuji-raf': ['raf'],
    'image/x-panasonic-rw2': ['rw2'],

    // Vector / CAD / metafiles
    'image/cgm': ['cgm'],
    'image/emf': ['emf'],
    'image/wmf': ['wmf'],
    'image/vnd.dwg': ['dwg'],
    'image/vnd.dxf': ['dxf'],
    'image/x-freehand': ['fh', 'fhc', 'fh4', 'fh5', 'fh7'],

    // Netpbm family
    'image/x-portable-anymap': ['pnm'],
    'image/x-portable-bitmap': ['pbm'],
    'image/x-portable-graymap': ['pgm'],
    'image/x-portable-pixmap': ['ppm'],

    // X11 / Unix
    'image/x-xbitmap': ['xbm'],
    'image/x-xpixmap': ['xpm'],
    'image/x-xwindowdump': ['xwd'],
    'image/x-cmu-raster': ['ras'],
    'image/x-rgb': ['rgb'],

    // Other legacy / niche
    'image/g3fax': ['g3'],
    'image/ief': ['ief'],
    'image/prs.btif': ['btif'],
    'image/sgi': ['sgi'],
    'image/vnd.djvu': ['djvu', 'djv'],
    'image/vnd.wap.wbmp': ['wbmp'],
    'image/x-pcx': ['pcx'],
    'image/x-tga': ['tga'],

    // Non-standard aliases seen in the wild
    'image/pjpeg': ['jpg', 'jpeg'],   // old IE progressive JPEG
    'image/x-png': ['png'],           // old IE PNG
    'image/x-ms-bmp': ['bmp'],
};

// Helper: MIME -> preferred extension
export function getImageExtension(mime: string) {
    const clean = mime.split(';')[0].trim().toLowerCase();
    return IMAGE_MIME_TYPES[clean]?.[0] ?? null;
}

export function formatByteSize(bytes: number, decimals: number = 2): string {
    if (!+bytes) return '0 Bytes'

    const k = 1024
    const dm = decimals < 0 ? 0 : decimals
    const sizes = ['Bytes', 'KiB', 'MiB', 'GiB', 'TiB', 'PiB', 'EiB', 'ZiB', 'YiB']

    const i = Math.floor(Math.log(bytes) / Math.log(k))

    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`
}