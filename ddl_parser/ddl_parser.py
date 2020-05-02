from ctypes import *
from sys import platform

class GoString(Structure):
    _fields_ = [("p", c_char_p), ("n", c_longlong)]

def get_parser_file():
    if platform == "linux" or platform == "linux2":
        return cdll.LoadLibrary("./ddl_parser_linux.so")
    elif platform == "darwin":
        return cdll.LoadLibrary("./ddl_parser.so")

def parse_ddl(sql):
    parser = get_parser_file()
    parser.Parse.argtypes = [GoString]
    parser.Parse.restype = c_char_p

    return parser.Parse(GoString(c_char_p(sql.encode('utf-8')), len(sql))).decode('utf-8')

print(parse_ddl('''
    CREATE TABLE example (
       id BIGINT UNSIGNED AUTO_INCREMENT,
       no_default INT NOT NULL,
       geom GEOMETRY NOT NULL SRID 4326,
       created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
       updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
       PRIMARY KEY (id),
       INDEX index_created_at (created_at),
       INDEX index_updated_at (updated_at),
       SPATIAL INDEX(geom)
    ) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
'''))

