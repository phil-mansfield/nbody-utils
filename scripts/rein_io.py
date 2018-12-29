from __future__ import print_function
import array
import struct
import numpy as np

class BinHHeader(object):
    pass

class BinHFile(object):
    def __init__(self, fname):
        self.fname = fname
        self._hd = BinHHeader()

        fp = open(self.fname, "rb")
        
        bin_h_header_text = fp.read(7*8)
        
        int_dtype = np.dtype(np.int64).newbyteorder("<")

        (self._hd._mode,
         self._hd._haloes, self._hd._column_num,
         self._hd._mass_column, self._hd._int_column_num,
         self._hd._names_len, self._hd._text_header_len) = struct.unpack(
             "<" + "q"*7, bin_h_header_text
        )

        int_column_bytes = fp.read(8*self._hd._int_column_num)
        self._hd.int_columns = np.frombuffer(
            int_column_bytes, dtype=int_dtype
        )

        column_names_bytes = fp.read(self._hd._names_len)
        self._hd.column_names = column_names_bytes.split("\n")
        self._hd.text_header = fp.read(self._hd._text_header_len)

        fp.close()

        self._name_lookup = { }
        for i in xrange(len(self._hd.column_names)):
            self._name_lookup[self._hd.column_names[i]] = i

        self._int_lookup = { }
        for col in self._hd.int_columns:
            self._int_lookup[col] = True


    def read_header(self):
        return self._hd
        
    def read_columns(self, columns):
        fp = open(self.fname, "rb")

        out = [None]*len(columns)

        int_dtype = np.dtype(np.int64).newbyteorder("<")
        float_dtype = np.dtype(np.float32).newbyteorder("<")

        for i in xrange(len(columns)):
            col = self._column_index(columns[i])
            if col in self._int_lookup:
                offset = self._int_offset(col)
                fp.seek(offset, 0)
                bytes = fp.read(8*self._hd._haloes)
                out[i] = np.frombuffer(bytes, dtype=int_dtype)
            else:
                offset = self._float_offset(col)
                fp.seek(offset, 0)
                bytes = fp.read(4*self._hd._haloes)
                out[i] = np.frombuffer(bytes, dtype=float_dtype)
            
        fp.close()
        
        return out

    def _column_index(self, column_id):
        if type(column_id) == str:
            if column_id in self._name_lookup:
                return self._name_lookup[column_id]
            else:
                raise ValueError("Column name '%s' not recognized" % column_id)
        elif column_id >= self._hd._column_num:
            raise ValueError("Column %d out of range: max column number is %d" %
                             (column_id, self._hd._column_num))
        else:
            return column_id

    def _int_offset(self, col):
        # TODO: vecorize
        idx = -1
        for i in xrange(len(self._hd.int_columns)):
            if self._hd.int_columns[i] == col:
                idx = i
                break
        if idx == -1: raise ValueError("Impossible")

        offset = (8*7 + # Binary header
                8*self._hd._int_column_num + # Int columns
                self._hd._names_len + # Columns names
                self._hd._text_header_len + # Text header
                idx*8*self._hd._haloes) # Other int columns

        return offset

    def _float_offset(self, col):
        # TODO: vecorize
        idx = col
        for i in xrange(len(self._hd.int_columns)):
            if self._hd.int_columns[i] < col: idx -= 1
            if self._hd.int_columns[i] == col: ValueError("Impossible")

        offset = (8*7 + # Binary header
                8*self._hd._int_column_num + # Int columns
                self._hd._names_len + # Columns names
                self._hd._text_header_len + # Text header
                8*self._hd._int_column_num*self._hd._haloes + # Int columns
                4*idx*self._hd._haloes) # Other float columns

        return offset

def main():
    bhf = BinHFile("tests/test.out")
    header = bhf.read_header()
    
    print(header.int_columns)
    print(header.column_names)
    print(header.text_header)

    data = bhf.read_columns(header.column_names[::1])

    for col in data:
        print(col)

if __name__ == "__main__": main()
