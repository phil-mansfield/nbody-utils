from __future__ import print_function, division
import numpy as np
import numpy.random as random
import struct

class ColumnFlag(object):
    def __init__(self, flag):
        self.name, self.size = [
            ("Float64", 8), # 0
            ("Float32", 4), # 1
            ("Int64", 8), # 2
            ("Int32", 4), # 3
            ("Int16", 2), # 4
            ("Int8", 1), # 5
            ("QFloat64", 8), # 6
            ("QFloat32", 4), # 7
            ("QFloat16", 2), # 8
            ("QFloat8", 1), # 9
            ("QLogFloat64", 8), # 10
            ("QLogFloat32", 4), # 11
            ("QLogFloat16", 2), # 12
            ("QLogFloat8", 1) # 13
        ][flag]

    def decode(self, fp, delta, key, haloes):
        if self.name == "Float64":
            dtype = np.dtype(np.float64).newbyteorder("<")
            return np.frombuffer(fp.read(haloes*8), dtype=dtype)
        elif self.name == "Float32":
            dtype = np.dtype(np.float32).newbyteorder("<")
            return np.frombuffer(fp.read(haloes*4), dtype=dtype)
        elif self.name == "Int64":
            dtype = np.dtype(np.int64).newbyteorder("<")
            idx = np.frombuffer(fp.read(haloes*8), dtype=dtype)
            return np.array(idx, dtype=int) + (key - np.iinfo(np.int64).min)
        elif self.name == "Int32":
            dtype = np.dtype(np.int32).newbyteorder("<")
            idx = np.frombuffer(fp.read(haloes*4), dtype=dtype)
            return np.array(idx, dtype=int) + (key - np.iinfo(np.int32).min)
        elif self.name == "Int16":
            dtype = np.dtype(np.int16).newbyteorder("<")
            idx = np.frombuffer(fp.read(haloes*2), dtype=dtype)
            return np.array(idx, dtype=int) + (key - np.iinfo(np.int16).min)
        elif self.name == "Int8":
            dtype = np.dtype(np.int8).newbyteorder("<")
            idx = np.frombuffer(fp.read(haloes), dtype=dtype)
            return np.array(idx, dtype=int) + (key - np.iinfo(np.int8).min)
        elif self.name == "QFloat64":
            idx = ColumnFlag(2).decode(fp, delta, key, haloes)
            return (idx + random.random(len(idx))) * delta
        elif self.name == "QFloat32":
            idx = ColumnFlag(3).decode(fp, delta, key, haloes)
            return (idx +random.random(len(idx))) * delta
        elif self.name == "QFloat16":
            idx = ColumnFlag(4).decode(fp, delta, key, haloes)
            return (idx + random.random(len(idx))) * delta
        elif self.name == "QFloat8":
            idx = ColumnFlag(5).decode(fp, delta, key, haloes)
            return (idx + random.random(len(idx))) * delta
        elif self.name == "QLogFloat64":
            return 10**ColumnFlag(6).decode(fp, delta, key, haloes)
        elif self.name == "QLogFloat32":
            return 10**ColumnFlag(7).decode(fp, delta, key, haloes)
        elif self.name == "QLogFloat16":
            lx = ColumnFlag(8).decode(fp, delta, key, haloes)
            return 10**lx
        elif self.name == "QLogFloat8":
            lx = ColumnFlag(9).decode(fp, delta, key, haloes)
            return 10**lx

        assert(0)

class BinhHeader(object):
    def __init__(self, fp):
        self.version = struct.unpack("<q", fp.read(8))[0]
        assert(self.version == 2)

        self.seed = struct.unpack("<q", fp.read(8))[0]
        self.columns = struct.unpack("<q", fp.read(8))[0]
        self.mass_column = struct.unpack("<q", fp.read(8))[0]
        self.blocks = struct.unpack("<q", fp.read(8))[0]
        self.text_header_length = struct.unpack("<q", fp.read(8))[0]
        self.text_column_names_length = struct.unpack("<q", fp.read(8))[0]
        self.is_sorted = struct.unpack("<q", fp.read(8))[0]
        self.min_mass = struct.unpack("<d", fp.read(8))[0]

        self.deltas = np.array(struct.unpack(
            "<%dd" % self.columns, fp.read(8*self.columns)
        ))
        self.column_skipped = np.array(
            struct.unpack("<%dB" % self.columns, fp.read(self.columns))
        )
        self.text_header = struct.unpack(
            "<%ds" % self.text_header_length, fp.read(self.text_header_length)
        )[0]
        self.column_names = struct.unpack(
            "<%ds" % self.text_column_names_length,
            fp.read(self.text_column_names_length)
        )[0].split(",")
        
        header_size = (8*9 + 8*self.columns + self.columns +
                       self.text_header_length + self.text_column_names_length)
        
        block_start = header_size
        self.block_haloes = np.zeros(self.blocks, dtype=int)
        self.block_flags = [None] * self.blocks
        self.block_keys = [None] * self.blocks
        self.data_offsets = np.zeros(self.blocks, dtype=int)
        for block in range(self.blocks):
            self.data_offsets[block] = block_start + 8*(1 + 2*self.columns)
            fp.seek(block_start, 0)

            self.block_haloes[block] = struct.unpack("<q", fp.read(8))[0]
            self.block_flags[block] = [
                ColumnFlag(flag) for flag in struct.unpack(
                    "<%dq" % self.columns, fp.read(8*self.columns),
                )]
            self.block_keys[block] = np.array(struct.unpack(
                "<%dq" % self.columns, fp.read(8*self.columns),
            ))

            block_start = self._next_block_header(block, block_start)

        self.haloes = np.sum(self.block_haloes)


    def _next_block_header(self, block, block_start):
        block_header_size = 8 + 8*self.columns + 8*self.columns
        column_size = 0
        for col in range(self.columns):
            if self.column_skipped[col] == 0:
                column_size += self.block_flags[block][col].size
        column_size *= self.block_haloes[block]

        return block_start + block_header_size + column_size
            

class Binh(object):
    def __init__(self, fname):
        self.fp = open(fname, "rb")
        self.hd = BinhHeader(self.fp)
        random.seed(self.hd.seed)

    def read_header(self):
        return self.hd

    def read(self, cols):
        out = [None] * len(cols)
        for i in range(len(cols)):
            out[i] = self.read_column(cols[i])
        return out

    def read_block(self, block, cols):
        out = [None] * len(cols)
        for i in range(len(cols)):
            out[i] = self.read_block_column(block, cols[i])
        return out

    def read_column(self, col):
        blocks = [None]*hd.blocks
        for block in range(hd.blocks):
            blocks[block] = self.read_block_column(block, col)
        return np.hstack(blocks)

    def read_block_column(self, block, col):
        col = self.int_col(col)

        data_offset = self.hd.data_offsets[block]
        for c in range(col):
            if self.hd.column_skipped[c] == 0:
                data_offset += (self.hd.block_haloes[block] *
                                self.hd.block_flags[block][c].size)

        self.fp.seek(data_offset, 0)

        flag = self.hd.block_flags[block][col]
        return flag.decode(
            self.fp, self.hd.deltas[col],
            self.hd.block_keys[block][col], self.hd.block_haloes[block]
        )
    
    def int_col(self, col):
        if type(col) == int: return col
        if type(col) == str:
            col = col.lower().strip()
            assert(col in self.hd.column_names)
            return self.hd.column_names.index(col)
