"""Embed a multi-size ICO as RT_ICON and RT_GROUP_ICON resources in a Go Windows EXE.

Uses only Python's standard library. Intended for the stripped, unsigned Go
PE32+ binary produced by build.bat. Run before code signing.
"""
import argparse
import struct
from pathlib import Path


def align(value, boundary):
    return (value + boundary - 1) // boundary * boundary


def embed_icon(exe_path, ico_path):
    ico = ico_path.read_bytes()
    reserved, kind, count = struct.unpack_from('<HHH', ico)
    if reserved or kind != 1 or not count:
        raise ValueError('Invalid ICO file')
    icons = []
    for i in range(count):
        w, h, palette, reserved, planes, bpp, length, pos = struct.unpack_from('<BBBBHHII', ico, 6 + 16 * i)
        if pos + length > len(ico):
            raise ValueError('Invalid ICO image data')
        icons.append((w, h, palette, planes, bpp, ico[pos:pos + length]))

    pe = bytearray(exe_path.read_bytes())
    if pe[:2] != b'MZ':
        raise ValueError('Not a Windows executable')
    peoff, = struct.unpack_from('<I', pe, 0x3c)
    if pe[peoff:peoff+4] != b'PE\0\0':
        raise ValueError('Invalid PE signature')
    header = peoff + 4
    sections, = struct.unpack_from('<H', pe, header + 2)
    optional_size, = struct.unpack_from('<H', pe, header + 16)
    optional = peoff + 24
    if struct.unpack_from('<H', pe, optional)[0] != 0x20b:
        raise ValueError('Expected a 64-bit Windows executable')
    section_align, file_align = struct.unpack_from('<II', pe, optional + 32)
    directory = optional + 112
    if struct.unpack_from('<II', pe, directory + 16) != (0, 0):
        raise ValueError('Executable already has a resource directory; rebuild the unsigned original first')
    sect_offset = optional + optional_size
    table = []
    for i in range(sections):
        offset = sect_offset + i * 40
        virtual_size, virtual_addr, raw_size, raw_offset = struct.unpack_from('<IIII', pe, offset + 8)
        table.append((virtual_size, virtual_addr, raw_size, raw_offset))
    new_hdr = sect_offset + sections * 40
    if new_hdr + 40 > min(s[3] for s in table):
        raise ValueError('Insufficient PE header space for a resource section')
    new_va = align(max(va + max(vs, rs) for vs, va, rs, ro in table), section_align)
    new_raw = align(max(ro + rs for vs, va, rs, ro in table), file_align)
    if len(pe) > new_raw:
        raise ValueError('Unexpected trailing binary data; refusing to overwrite')

    data = bytearray()
    def reserve(size):
        at = len(data)
        data.extend(b'\0' * size)
        return at
    def pad():
        data.extend(b'\0' * ((-len(data)) % 4))
    root = reserve(16)
    root_icon = reserve(8)
    root_group = reserve(8)
    icon_dir = reserve(16)
    icon_entries = [reserve(8) for _ in icons]
    group_dir = reserve(16)
    group_entry = reserve(8)
    languages = []
    for _ in icons:
        languages.append((reserve(16), reserve(8)))
    group_lang = reserve(16)
    group_lang_entry = reserve(8)
    image_entries = [reserve(16) for _ in icons]
    group_data_entry = reserve(16)
    pad()
    images = []
    for icon in icons:
        at = len(data)
        data.extend(icon[5]); pad()
        images.append((at, len(icon[5])))
    group = bytearray(struct.pack('<HHH', 0, 1, len(icons)))
    for i, (w, h, palette, planes, bpp, payload) in enumerate(icons, 1):
        group.extend(struct.pack('<BBBBHHIH', w, h, palette, 0, planes, bpp, len(payload), i))
    group_at = len(data)
    data.extend(group); pad()
    struct.pack_into('<IIHHHH', data, root, 0, 0, 0, 0, 0, 2)
    struct.pack_into('<II', data, root_icon, 3, 0x80000000 | icon_dir)
    struct.pack_into('<II', data, root_group, 14, 0x80000000 | group_dir)
    struct.pack_into('<IIHHHH', data, icon_dir, 0, 0, 0, 0, 0, len(icons))
    for i, entry in enumerate(icon_entries, 1):
        struct.pack_into('<II', data, entry, i, 0x80000000 | languages[i-1][0])
    struct.pack_into('<IIHHHH', data, group_dir, 0, 0, 0, 0, 0, 1)
    struct.pack_into('<II', data, group_entry, 1, 0x80000000 | group_lang)
    for i, (ldir, lentry) in enumerate(languages):
        struct.pack_into('<IIHHHH', data, ldir, 0, 0, 0, 0, 0, 1)
        struct.pack_into('<II', data, lentry, 1033, image_entries[i])
        struct.pack_into('<IIII', data, image_entries[i], new_va + images[i][0], images[i][1], 0, 0)
    struct.pack_into('<IIHHHH', data, group_lang, 0, 0, 0, 0, 0, 1)
    struct.pack_into('<II', data, group_lang_entry, 1033, group_data_entry)
    struct.pack_into('<IIII', data, group_data_entry, new_va + group_at, len(group), 0, 0)

    raw_size = align(len(data), file_align)
    pe.extend(b'\0' * (new_raw - len(pe)))
    pe.extend(data)
    pe.extend(b'\0' * (raw_size - len(data)))
    struct.pack_into('<8sIIIIIIHHI', pe, new_hdr, b'.rsrc\0\0\0', len(data), new_va, raw_size, new_raw, 0, 0, 0, 0, 0x40000040)
    struct.pack_into('<H', pe, header + 2, sections + 1)
    struct.pack_into('<I', pe, optional + 56, align(new_va + len(data), section_align))
    struct.pack_into('<II', pe, directory + 16, new_va, len(data))
    struct.pack_into('<I', pe, optional + 64, 0)  # checksum not required for unsigned EXE
    exe_path.write_bytes(pe)
    print(f'Embedded {len(icons)} icon sizes into {exe_path.name}')


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('exe', type=Path)
    parser.add_argument('ico', type=Path)
    args = parser.parse_args()
    embed_icon(args.exe, args.ico)
