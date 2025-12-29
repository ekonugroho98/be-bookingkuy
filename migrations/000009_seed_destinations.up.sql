-- Seed Indonesian destinations for initial testing
-- This data can be replaced later by syncing from HotelBeds API

INSERT INTO destinations (code, name, country_code, country_name, type, latitude, longitude) VALUES
-- Major Cities
('JKT', 'Jakarta', 'ID', 'Indonesia', 'CITY', -6.2088, 106.8456),
('SUB', 'Surabaya', 'ID', 'Indonesia', 'CITY', -7.2575, 112.7521),
('BDO', 'Bandung', 'ID', 'Indonesia', 'CITY', -6.9175, 107.6191),
('DPS', 'Denpasar', 'ID', 'Indonesia', 'CITY', -8.6705, 115.2126),
('MDN', 'Medan', 'ID', 'Indonesia', 'CITY', 3.5952, 98.6722),
('UPG', 'Makassar', 'ID', 'Indonesia', 'CITY', -5.1477, 119.4328),
('YIA', 'Yogyakarta', 'ID', 'Indonesia', 'CITY', -7.7956, 110.3695),
('SMG', 'Semarang', 'ID', 'Indonesia', 'CITY', -6.9667, 110.4167),
('BKS', 'Bekasi', 'ID', 'Indonesia', 'CITY', -6.2349, 106.9763),
('TNG', 'Tangerang', 'ID', 'Indonesia', 'CITY', -6.1783, 106.6319),

-- Popular Tourist Destinations
('BTW', 'Batu', 'ID', 'Indonesia', 'CITY', -7.8667, 112.5333),
('MLG', 'Malang', 'ID', 'Indonesia', 'CITY', -7.9797, 112.6304),
 ('BJL', 'Banjarmasin', 'ID', 'Indonesia', 'CITY', -3.3167, 114.5900),
('BPN', 'Balikpapan', 'ID', 'Indonesia', 'CITY', -1.2654, 116.8311),
 ('PKU', 'Pekanbaru', 'ID', 'Indonesia', 'CITY', 0.5071, 101.4479),
 ('JOG', 'Jogjakarta', 'ID', 'Indonesia', 'CITY', -7.7971, 110.3705),
 ('LMP', 'Lampung', 'ID', 'Indonesia', 'CITY', -5.4292, 105.2612),
 ('PQL', 'Palembang', 'ID', 'Indonesia', 'CITY', -2.9761, 104.7604),
 ('MJK', 'Manado', 'ID', 'Indonesia', 'CITY', 1.4748, 124.8451),
 ('PLY', 'Palu', 'ID', 'Indonesia', 'CITY', -0.9000, 119.8667),

-- Bali Destinations
('GIL', 'Gili Islands', 'ID', 'Indonesia', 'CITY', -8.3495, 116.0528),
('UBU', 'Ubud', 'ID', 'Indonesia', 'CITY', -8.5069, 115.2625),
('KUT', 'Kuta', 'ID', 'Indonesia', 'CITY', -8.7225, 115.1689),
('SAN', 'Sanur', 'ID', 'Indonesia', 'CITY', -8.7222, 115.2633),
('SEMI', 'Seminyak', 'ID', 'Indonesia', 'CITY', -8.6909, 115.1583),
('CANG', 'Canggu', 'ID', 'Indonesia', 'CITY', -8.6589, 115.1286),
('Lovina', 'Lovina', 'ID', 'Indonesia', 'CITY', -8.1833, 115.0333),
('Amed', 'Amed', 'ID', 'Indonesia', 'CITY', -8.2917, 115.5750),
('NUSA', 'Nusa Dua', 'ID', 'Indonesia', 'CITY', -8.7870, 115.2256),
('JIMB', 'Jimbaran', 'ID', 'Indonesia', 'CITY', -8.7467, 115.1753),

-- West Java Destinations
('BOG', 'Bogor', 'ID', 'Indonesia', 'CITY', -6.5944, 106.7892),
('CIRE', 'Cirebon', 'ID', 'Indonesia', 'CITY', -6.7268, 108.5540),
('SUK', 'Sukabumi', 'ID', 'Indonesia', 'CITY', -6.9272, 106.9333),
('TAS', 'Tasikmalaya', 'ID', 'Indonesia', 'CITY', -7.3547, 108.2333),
('GAR', 'Garut', 'ID', 'Indonesia', 'CITY', -7.2517, 107.8869),
('CIAN', 'Cianjur', 'ID', 'Indonesia', 'CITY', -6.8167, 107.1417),
('MAJ', 'Majalengka', 'ID', 'Indonesia', 'CITY', -6.7389, 108.2292),
('KUN', 'Kuningan', 'ID', 'Indonesia', 'CITY', -6.9792, 108.4750),
('SUM', 'Sumedang', 'ID', 'Indonesia', 'CITY', -6.8417, 107.9219),
('PANG', 'Pangandaran', 'ID', 'Indonesia', 'CITY', -7.6833, 108.4833),

-- Central Java Destinations
('TEG', 'Tegal', 'ID', 'Indonesia', 'CITY', -6.9744, 109.1381),
('PEK', 'Pekalongan', 'ID', 'Indonesia', 'CITY', -6.8886, 109.6725),
('KUD', 'Kudus', 'ID', 'Indonesia', 'CITY', -6.8067, 110.8439),
('PATI', 'Pati', 'ID', 'Indonesia', 'CITY', -6.7444, 111.0364),
('KEND', 'Kendal', 'ID', 'Indonesia', 'CITY', -7.0086, 110.2069),
('MAGEL', 'Magelang', 'ID', 'Indonesia', 'CITY', -7.4675, 110.2178),
('KLAT', 'Klaten', 'ID', 'Indonesia', 'CITY', -7.6056, 110.5939),
('SOLO', 'Solo', 'ID', 'Indonesia', 'CITY', -7.5617, 110.8314),
('WON', 'Wonosobo', 'ID', 'Indonesia', 'CITY', -7.3636, 109.8917),
('BANY', 'Banyumas', 'ID', 'Indonesia', 'CITY', -7.4000, 109.3167),

-- East Java Destinations
('PROB', 'Probolinggo', 'ID', 'Indonesia', 'CITY', -7.7333, 113.2167),
('PASU', 'Pasuruan', 'ID', 'Indonesia', 'CITY', -7.6444, 112.8967),
('MOJO', 'Mojokerto', 'ID', 'Indonesia', 'CITY', -7.4717, 112.4389),
('JOMB', 'Jombang', 'ID', 'Indonesia', 'CITY', -7.5528, 112.2336),
('NGAN', 'Nganjuk', 'ID', 'Indonesia', 'CITY', -7.6017, 111.9011),
('MADI', 'Madiun', 'ID', 'Indonesia', 'CITY', -7.6297, 111.5281),
('KED', 'Kediri', 'ID', 'Indonesia', 'CITY', -7.8525, 112.0158),
('BLIT', 'Blitar', 'ID', 'Indonesia', 'CITY', -8.0969, 112.1500),
('TRENG', 'Trenggalek', 'ID', 'Indonesia', 'CITY', -8.0653, 111.6989),
('TULUNG', 'Tulungagung', 'ID', 'Indonesia', 'CITY', -8.0689, 111.9147),

-- Sumatra Destinations
('PDG', 'Padang', 'ID', 'Indonesia', 'CITY', -0.9471, 100.4172),
('BATAM', 'Batam', 'ID', 'Indonesia', 'CITY', 1.1306, 103.9747),
('PEKAN', 'Pekanbaru', 'ID', 'Indonesia', 'CITY', 0.5071, 101.4479),
('JAMBI', 'Jambi', 'ID', 'Indonesia', 'CITY', -1.5906, 103.6081),
('BENGK', 'Bengkulu', 'ID', 'Indonesia', 'CITY', -3.8000, 102.2667),
('LHOK', 'Banda Aceh', 'ID', 'Indonesia', 'CITY', 5.5500, 95.3167),
('DUMA', 'Dumai', 'ID', 'Indonesia', 'CITY', 1.6967, 101.4450),
('TEB', 'Tebing Tinggi', 'ID', 'Indonesia', 'CITY', 3.4167, 99.1500),
('BIN', 'Binjai', 'ID', 'Indonesia', 'CITY', 3.6000, 98.4833),
('PADAN', 'Padang Sidempuan', 'ID', 'Indonesia', 'CITY', 1.3833, 99.2667),

-- Kalimantan Destinations
('SAMBAS', 'Sambas', 'ID', 'Indonesia', 'CITY', 1.3833, 109.2833),
('SANG', 'Sangatta', 'ID', 'Indonesia', 'CITY', -0.5333, 116.5500),
('BONT', 'Bontang', 'ID', 'Indonesia', 'CITY', 0.2167, 117.4667),
('SAMAR', 'Samarinda', 'ID', 'Indonesia', 'CITY', -0.4956, 117.1436),

-- Sulawesi Destinations
('GOR', 'Gorontalo', 'ID', 'Indonesia', 'CITY', 0.5417, 123.0167),
('KENDA', 'Kendari', 'ID', 'Indonesia', 'CITY', -3.9822, 122.5150),
('POS', 'Poso', 'ID', 'Indonesia', 'CITY', -1.4000, 120.7500),
('PALU', 'Palu', 'ID', 'Indonesia', 'CITY', -0.9000, 119.8667),
('MANADO', 'Manado', 'ID', 'Indonesia', 'CITY', 1.4748, 124.8451),
('BITUNG', 'Bitung', 'ID', 'Indonesia', 'CITY', 1.4333, 125.1833),
('TOMO', 'Tomohon', 'ID', 'Indonesia', 'CITY', 1.3167, 124.8333),

-- Papua & Maluku
('AMB', 'Ambon', 'ID', 'Indonesia', 'CITY', -3.6956, 128.1817),
('TERN', 'Ternate', 'ID', 'Indonesia', 'CITY', 0.8783, 127.3139),
('TID', 'Tidore', 'ID', 'Indonesia', 'CITY', 0.8333, 127.4000),
('JAYA', 'Jayapura', 'ID', 'Indonesia', 'CITY', -2.5337, 140.7181),
('SORO', 'Sorong', 'ID', 'Indonesia', 'CITY', -0.8917, 131.2883),
('MANOK', 'Manokwari', 'ID', 'Indonesia', 'CITY', -0.8667, 134.0833),
('BIAK', 'Biak', 'ID', 'Indonesia', 'CITY', -1.1833, 136.0833),

-- Popular Tourist Islands
('BALI', 'Bali Island', 'ID', 'Indonesia', 'CITY', -8.3405, 115.0920),
('GILI', 'Gili Trawangan', 'ID', 'Indonesia', 'CITY', -8.3495, 116.0528),
('KOMO', 'Komodo Island', 'ID', 'Indonesia', 'CITY', -8.5500, 119.4667),
('RAJA', 'Raja Ampat', 'ID', 'Indonesia', 'CITY', -0.2333, 130.5000),
('WAKA', 'Wakatobi', 'ID', 'Indonesia', 'CITY', -5.3167, 123.5833),
('BROM', 'Mount Bromo', 'ID', 'Indonesia', 'CITY', -7.9425, 112.9500),
('IJEN', 'Kawah Ijen', 'ID', 'Indonesia', 'CITY', -8.0583, 114.2417),
('BORO', 'Borobudur', 'ID', 'Indonesia', 'CITY', -7.6075, 110.2036),
('PRAM', 'Prambanan', 'ID', 'Indonesia', 'CITY', -7.7517, 110.4914),
('UBUD', 'Ubud', 'ID', 'Indonesia', 'CITY', -8.5069, 115.2625)
ON CONFLICT (code) DO NOTHING;
