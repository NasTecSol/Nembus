--
-- PostgreSQL database dump
--

\restrict c1HeHpRg3Idak5CAUN6FHDDG915sgtrinT5hLo7Gnzu9bDpfPfKEdpkEJwfTOii

-- Dumped from database version 16.14
-- Dumped by pg_dump version 18.1

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Data for Name: audit_logs; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.audit_logs (id, organization_id, table_name, record_id, action, old_values, new_values, changed_fields, performed_by, ip_address, user_agent, session_id, created_at) FROM stdin;
\.


--
-- Data for Name: brands; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.brands (id, name, code, is_active, metadata, created_at, updated_at) FROM stdin;
1	Almarai	ALMARAI	t	{"country": "Saudi Arabia", "category": "Dairy"}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
2	Nadec	NADEC	t	{"country": "Saudi Arabia", "category": "Dairy"}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
3	Al-Safi Danone	ALSAFI	t	{"country": "Saudi Arabia", "category": "Dairy"}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
4	Americana	AMERICANA	t	{"country": "Kuwait", "category": "Food"}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
5	Nestlé	NESTLE	t	{"country": "International", "category": "Food & Beverages"}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
6	Coca-Cola	COCACOLA	t	{"country": "International", "category": "Beverages"}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
7	PepsiCo	PEPSI	t	{"country": "International", "category": "Beverages"}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
8	Al-Watania	WATANIA	t	{"country": "Saudi Arabia", "category": "Poultry"}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
9	Sunbulah	SUNBULAH	t	{"country": "Saudi Arabia", "category": "Frozen Foods"}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
10	California Garden	CALGARDEN	t	{"country": "Saudi Arabia", "category": "Canned Foods"}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
11	Rabea	RABEA	t	{"country": "Saudi Arabia", "category": "Tea"}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
12	Lipton	LIPTON	t	{"country": "International", "category": "Tea"}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
13	Nescafe	NESCAFE	t	{"country": "International", "category": "Coffee"}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
14	Lavazza	LAVAZZA	t	{"country": "Italy", "category": "Coffee"}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
15	Dettol	DETTOL	t	{"country": "International", "category": "Personal Care"}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
17	Ariel	ARIEL	t	{"country": "International", "category": "Household"}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
18	Persil	PERSIL	t	{"country": "International", "category": "Household"}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
19	Dove	DOVE	t	{"country": "International", "category": "Personal Care"}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
20	Lux	LUX	t	{"country": "International", "category": "Personal Care"}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
21	Palmolive	PALMOLIVE	t	{"country": "International", "category": "Personal Care"}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
22	Finish	FINISH	t	{"country": "International", "category": "Household"}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
23	Samsung	SAMSUNG	t	{"country": "South Korea", "category": "Electronics"}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
24	LG	LG	t	{"country": "South Korea", "category": "Electronics"}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
25	Panasonic	PANASONIC	t	{"country": "Japan", "category": "Electronics"}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
26	Philips	PHILIPS	t	{"country": "Netherlands", "category": "Electronics"}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
16	Tide	TIDE	t	{"country": "International", "category": "Household", "website_url": "https://www.toyota.com"}	2026-07-18 07:58:31.573111	2026-07-24 06:24:35.865095
\.


--
-- Data for Name: cart_activity_log; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.cart_activity_log (id, cart_id, organization_id, activity_type, description, performed_by_user_id, ip_address, user_agent, old_value, new_value, created_at) FROM stdin;
1	786d75ea-064f-4ec6-9f1e-15a9cd122adc	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-20 08:21:31.507485
2	1b665464-53d1-4f00-9076-ccc0b9d8d058	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-20 08:31:16.39892
3	56324b30-f1ce-48c0-aea8-be446dff854c	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-20 11:03:02.714912
4	b6933e63-f839-4d53-a4da-501b113d878a	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-20 11:43:31.225812
5	3dfb4094-071e-40a5-bda1-c6e2159e0c7c	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-21 07:11:31.105103
6	ea501529-ce47-4e5f-898a-6c9e2304c995	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-21 07:12:23.889965
7	37447439-4024-451a-8807-b8cfca010487	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-21 07:12:48.86144
8	2a7a15e0-9167-4524-abbd-bfd138bbaa97	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-22 06:26:08.053801
9	def65158-beb8-4300-83d5-fba82480ba3e	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-22 07:37:06.207043
10	d8d59145-07e8-467c-b016-9555c1af7a88	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-22 07:37:39.263321
11	baccf90f-7f2c-4974-b935-5754c1f8f28a	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-22 07:38:22.782073
12	6912df30-1a48-4569-af67-9885dccd1109	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-22 09:05:40.908668
13	c03314a4-0319-467c-81c7-32b20fe42417	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-22 09:52:19.48296
14	c03314a4-0319-467c-81c7-32b20fe42417	1	status_changed	Cart status changed from converted to active	\N	\N	\N	{"status": "converted"}	{"status": "active"}	2026-07-22 09:52:27.305781
15	c03314a4-0319-467c-81c7-32b20fe42417	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-22 09:54:51.07443
16	161c21c8-61c7-4d8c-b4a0-3c9f7e760e3e	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-22 10:01:20.554113
17	161c21c8-61c7-4d8c-b4a0-3c9f7e760e3e	1	status_changed	Cart status changed from converted to active	\N	\N	\N	{"status": "converted"}	{"status": "active"}	2026-07-22 10:01:40.748179
18	161c21c8-61c7-4d8c-b4a0-3c9f7e760e3e	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-22 10:02:26.789285
19	6bbea337-0a12-45a9-afb8-0e8eaeeaf9da	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-22 10:02:51.748435
20	156c1666-36f7-4a09-92e5-22dc40cbeba8	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-22 10:05:41.923776
21	156c1666-36f7-4a09-92e5-22dc40cbeba8	1	status_changed	Cart status changed from converted to active	\N	\N	\N	{"status": "converted"}	{"status": "active"}	2026-07-22 10:05:51.939357
22	156c1666-36f7-4a09-92e5-22dc40cbeba8	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-22 10:07:14.531335
23	156c1666-36f7-4a09-92e5-22dc40cbeba8	1	status_changed	Cart status changed from converted to active	\N	\N	\N	{"status": "converted"}	{"status": "active"}	2026-07-22 10:07:28.181571
24	f3bc32e2-0d76-4bb6-9ed8-7f02b3021566	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-22 13:11:05.655009
25	6f23c0ff-e3eb-4ba6-9996-2225ed1afb14	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-22 13:19:08.042946
26	6f23c0ff-e3eb-4ba6-9996-2225ed1afb14	1	status_changed	Cart status changed from converted to active	\N	\N	\N	{"status": "converted"}	{"status": "active"}	2026-07-22 13:19:11.187794
27	4985babb-0bfa-4532-b61e-63a989fa9ee5	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-23 06:18:29.318977
28	93837ac6-d0d5-4bb6-b6fe-aae9278f91c1	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-23 06:28:34.485405
29	e3f11cd0-804c-4c79-95e0-7b910cfacb6d	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-23 06:37:55.624729
30	ca973e0c-c7d6-4a65-a74e-895089e0527a	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-23 08:31:04.30609
31	0b90118f-6987-47c4-bc66-3325bd219b63	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-23 10:54:04.965803
32	a0e2078d-2e1b-48a5-896f-ebc458458dc5	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-23 11:07:35.557081
33	b4ab7b6b-7e3c-43a9-b297-7e4950fd3b96	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-23 11:16:48.139645
34	d8c30107-0eb2-4855-bb6a-51cd53505b5d	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-23 11:24:59.685238
35	c269c10a-80e9-41ee-a08b-9c5360c2e4d1	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-23 11:42:46.148176
36	bfd00d5b-5d48-491b-ade1-b18de6afcb68	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-24 09:14:58.90756
37	2eb60586-650e-4927-a018-945a436f01a3	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-27 10:47:45.066852
38	c62c4d60-b9e9-4c32-9c09-413b37bed03e	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-27 11:06:26.237358
39	a2eebeb3-662c-46df-82ef-e6ad81058c5b	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-27 11:11:19.925869
40	14edb196-5851-4397-b37d-66322afe2a2d	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-27 11:42:20.346292
41	e1dad131-4aa2-487a-b854-e002b43febe4	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-27 11:56:46.137556
42	334043fe-9a22-4e2d-b316-8b780d6717dd	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-27 12:05:02.734041
43	339d0029-bbdf-49c3-8fc1-cadba066f020	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-27 12:06:45.369439
44	25f07dc1-0c1b-4fc4-a9a5-da5398c129a7	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-28 08:35:54.498196
45	7a73337e-1c3b-4c14-9cd9-9e449851d7ee	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-28 08:37:14.651939
46	62161eda-60af-4495-a576-c6d352e0b040	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-30 08:02:53.772947
47	252baaaf-5ae9-4d3a-84fe-3e0878ca2b3f	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-30 08:10:35.372733
48	05664a1e-fd9e-4045-81b4-fd5cb0d4110c	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-30 10:59:28.61012
49	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-30 11:04:16.88959
50	aed43720-0c89-4be2-bec4-cbba48662f56	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-30 11:08:54.534423
51	a6a8a674-e57e-4809-9e86-b7d517ef7d4f	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-31 09:02:44.031443
52	d397fad3-1a15-4db2-a6db-bec5bf2b9297	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-31 09:05:05.107532
53	8f1bca0d-e6d0-434e-9895-8a68343bd13f	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-31 09:11:02.466443
54	3469e4f7-1aef-475b-951b-2cd3ec962029	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-31 09:13:36.240484
55	3b93a016-e219-4b71-a2a8-9ef8886d2aae	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-07-31 10:19:10.720094
56	c37191dd-8749-477b-bbe3-402c2e09a31d	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-08-01 06:18:24.44564
57	b627c8a2-bc74-4bb8-9b28-0790aab1266a	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-08-01 06:26:48.164436
58	de7f9185-027d-4156-bf2c-7b03cd948e88	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-08-01 06:47:31.587683
59	7750e699-09d6-45b6-a904-e408f9dabaef	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-08-03 05:31:22.813367
60	b014c45e-d107-4546-9852-0ceafe1d6282	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-08-03 09:55:00.656239
61	35aa084d-d355-48f6-88d7-8817ee31408b	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-08-03 10:34:21.520275
62	dabc964f-45e9-4234-a23d-4e2fddcdaaee	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-08-03 10:36:05.368178
63	9f357046-da05-423c-8753-9c3031e83a44	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-08-03 10:36:47.658755
64	73419d52-4182-413f-a4da-14cc77155a65	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-08-03 11:39:23.332308
65	4620d829-c1f1-4885-9686-3bf3796b3cbb	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-08-03 11:41:39.519443
66	a0f2df8d-1907-41dd-927d-749d3e09730b	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-08-11 11:24:31.117197
67	ecaed928-ee99-437d-8e8d-e46d4597aff8	1	status_changed	Cart status changed from active to converted	\N	\N	\N	{"status": "active"}	{"status": "converted"}	2026-08-11 11:46:45.409218
\.


--
-- Data for Name: cart_items; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.cart_items (id, cart_id, organization_id, product_id, product_variant_id, quantity, uom_id, unit_price, discount_amount, tax_amount, line_total, price_list_id, tax_category_id, batch_number, serial_number, customization_details, notes, metadata, added_at, updated_at) FROM stdin;
722fcf57-1ee8-45df-8a2d-607c336c779f	d8c30107-0eb2-4855-bb6a-51cd53505b5d	1	34	\N	1.000	11	24.95	\N	3.74	28.69	1	1	\N	\N	\N	\N	\N	2026-07-20 07:47:47.005816	2026-07-20 07:47:47.005816
e2c58ffc-be94-4321-a335-ae6881f96440	b4ab7b6b-7e3c-43a9-b297-7e4950fd3b96	1	34	\N	1.000	11	24.95	\N	3.74	28.69	1	1	\N	\N	\N	\N	\N	2026-07-20 07:49:42.057309	2026-07-20 07:49:42.057309
c780d140-f5fc-4138-81ee-a69bc519634f	b4ab7b6b-7e3c-43a9-b297-7e4950fd3b96	1	26	\N	1.000	8	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-07-20 07:49:43.896343	2026-07-20 07:49:43.896343
0b36a106-20d2-481e-b440-77f3f0d4facc	b4ab7b6b-7e3c-43a9-b297-7e4950fd3b96	1	40	\N	1.000	4	48.50	\N	7.27	55.77	1	1	\N	\N	\N	\N	\N	2026-07-20 07:51:19.358751	2026-07-20 07:51:19.358751
8104fb44-17a3-41e2-a9d6-e5f8b686bf53	b4ab7b6b-7e3c-43a9-b297-7e4950fd3b96	1	40	\N	3.000	4	48.50	\N	21.82	167.32	1	1	\N	\N	\N	\N	\N	2026-07-20 08:20:59.637092	2026-07-20 08:20:59.637092
2b25b8a4-58a6-4a4e-8743-4313defebab1	786d75ea-064f-4ec6-9f1e-15a9cd122adc	1	8	\N	10.000	1	18.00	\N	27.00	207.00	1	1	\N	\N	\N	\N	\N	2026-07-20 08:21:26.346507	2026-07-20 08:21:26.346507
aa2624f2-811b-4d23-94ba-04bb73c25018	786d75ea-064f-4ec6-9f1e-15a9cd122adc	1	2	\N	1.000	3	8.50	\N	1.27	9.78	1	1	\N	\N	\N	\N	\N	2026-07-20 08:21:29.199133	2026-07-20 08:21:29.199133
ce61dffa-1503-4120-ba36-35b41c07f72e	1b665464-53d1-4f00-9076-ccc0b9d8d058	1	40	\N	6.000	4	48.50	\N	43.65	334.65	1	1	\N	\N	\N	\N	\N	2026-07-20 08:31:04.181736	2026-07-20 08:31:04.181736
5b67ccf5-eaab-48ea-9d12-f99693ef79b8	1b665464-53d1-4f00-9076-ccc0b9d8d058	1	23	\N	1.000	3	19.95	\N	2.99	22.94	1	1	\N	\N	\N	\N	\N	2026-07-20 08:31:05.956497	2026-07-20 08:31:05.956497
3d3a9eb6-140f-42e3-af07-0c317ed8b47a	1b665464-53d1-4f00-9076-ccc0b9d8d058	1	7	\N	1.000	10	15.50	\N	2.32	17.82	1	1	\N	\N	\N	\N	\N	2026-07-20 08:31:13.054929	2026-07-20 08:31:13.054929
0ff0fa96-74c1-40d3-8855-83aea79c93ae	56324b30-f1ce-48c0-aea8-be446dff854c	1	34	\N	1.000	11	24.95	\N	3.74	28.69	1	1	\N	\N	\N	\N	\N	2026-07-20 11:02:57.967034	2026-07-20 11:02:57.967034
d858b0af-7817-4e32-b702-ad546ceafc29	56324b30-f1ce-48c0-aea8-be446dff854c	1	26	\N	1.000	8	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-07-20 11:02:58.487584	2026-07-20 11:02:58.487584
2b1e0968-ed14-4bf6-a07f-836aa7bb47b1	56324b30-f1ce-48c0-aea8-be446dff854c	1	6	\N	1.000	10	12.95	\N	1.94	14.89	1	1	\N	\N	\N	\N	\N	2026-07-20 11:02:59.969627	2026-07-20 11:02:59.969627
8a6bd0ef-3ed4-49bc-a8ef-d2b5f49ef7d4	b6933e63-f839-4d53-a4da-501b113d878a	1	34	\N	1.000	11	24.95	\N	3.74	28.69	1	1	\N	\N	\N	\N	\N	2026-07-20 11:43:27.817493	2026-07-20 11:43:27.817493
5123066d-f589-4f7f-b289-810bcc579699	b6933e63-f839-4d53-a4da-501b113d878a	1	35	\N	1.000	10	3.50	\N	0.53	4.03	1	1	\N	\N	\N	\N	\N	2026-07-20 11:43:28.681905	2026-07-20 11:43:28.681905
b860c88f-8a22-44c4-9a8b-a2a99c291d32	3dfb4094-071e-40a5-bda1-c6e2159e0c7c	1	33	\N	1.000	10	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-07-21 07:11:27.611127	2026-07-21 07:11:27.611127
39d3a7b7-26ee-4f43-8f17-405867fed7f1	3dfb4094-071e-40a5-bda1-c6e2159e0c7c	1	26	\N	1.000	8	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-07-21 07:11:28.172326	2026-07-21 07:11:28.172326
de611272-c524-4222-b889-841729b37310	3dfb4094-071e-40a5-bda1-c6e2159e0c7c	1	34	\N	1.000	11	24.95	\N	3.74	28.69	1	1	\N	\N	\N	\N	\N	2026-07-21 07:11:29.106134	2026-07-21 07:11:29.106134
cdb79d0c-b648-4a0f-b2a0-b2dd08fa32bd	ea501529-ce47-4e5f-898a-6c9e2304c995	1	34	\N	1.000	11	24.95	\N	3.74	28.69	1	1	\N	\N	\N	\N	\N	2026-07-21 07:11:47.214628	2026-07-21 07:11:47.214628
09a26220-b641-4323-96d6-8cb917c7575b	ea501529-ce47-4e5f-898a-6c9e2304c995	1	26	\N	1.000	8	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-07-21 07:11:48.140653	2026-07-21 07:11:48.140653
d79f0dd0-e87f-4bf3-ace6-29129795e1e3	ea501529-ce47-4e5f-898a-6c9e2304c995	1	6	\N	1.000	10	12.95	\N	1.94	14.89	1	1	\N	\N	\N	\N	\N	2026-07-21 07:11:48.970178	2026-07-21 07:11:48.970178
47e9b381-0e1b-4d2b-ba6e-d36d43f85e4f	37447439-4024-451a-8807-b8cfca010487	1	35	\N	1.000	10	3.50	\N	0.53	4.03	1	1	\N	\N	\N	\N	\N	2026-07-21 07:12:42.52723	2026-07-21 07:12:42.52723
18782af6-6c67-4571-b0ef-82f75102e4ed	37447439-4024-451a-8807-b8cfca010487	1	27	\N	1.000	8	8.95	\N	1.34	10.29	1	1	\N	\N	\N	\N	\N	2026-07-21 07:12:42.966924	2026-07-21 07:12:42.966924
7318d81f-312c-4ead-8569-e18d2fb8f80c	37447439-4024-451a-8807-b8cfca010487	1	40	\N	6.000	4	48.50	\N	43.65	334.65	1	1	\N	\N	\N	\N	\N	2026-07-21 07:12:46.320548	2026-07-21 07:12:46.320548
b8d925a9-fd19-4613-91b1-11f5aa48d692	2a7a15e0-9167-4524-abbd-bfd138bbaa97	1	33	\N	1.000	10	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-07-22 06:26:02.051585	2026-07-22 06:26:02.051585
52130d8a-9b78-4808-9be6-a3e1c0b633ee	2a7a15e0-9167-4524-abbd-bfd138bbaa97	1	35	\N	1.000	10	3.50	\N	0.53	4.03	1	1	\N	\N	\N	\N	\N	2026-07-22 06:26:03.078759	2026-07-22 06:26:03.078759
fdbcacdb-527b-4d0b-93f6-c6f11bb176d2	2a7a15e0-9167-4524-abbd-bfd138bbaa97	1	6	\N	1.000	10	12.95	\N	1.94	14.89	1	1	\N	\N	\N	\N	\N	2026-07-22 06:26:05.42092	2026-07-22 06:26:05.42092
ea127c41-ea5e-4531-a5a8-83e10e878c7e	def65158-beb8-4300-83d5-fba82480ba3e	1	34	\N	1.000	11	24.95	\N	3.74	28.69	1	1	\N	\N	\N	\N	\N	2026-07-22 07:37:03.351921	2026-07-22 07:37:03.351921
654dca61-cbfd-46cb-a9a1-dc57b6c273a0	def65158-beb8-4300-83d5-fba82480ba3e	1	6	\N	1.000	10	12.95	\N	1.94	14.89	1	1	\N	\N	\N	\N	\N	2026-07-22 07:37:04.43384	2026-07-22 07:37:04.43384
dca20123-d7c5-4ac6-a73d-934730de10c9	d8d59145-07e8-467c-b016-9555c1af7a88	1	40	\N	6.000	4	48.50	\N	43.65	334.65	1	1	\N	\N	\N	\N	\N	2026-07-22 07:37:31.434656	2026-07-22 07:37:31.434656
f52f219d-8004-4ac5-82b7-77631d5fea85	baccf90f-7f2c-4974-b935-5754c1f8f28a	1	21	\N	5.000	2	85.00	\N	63.75	488.75	1	1	\N	\N	\N	\N	\N	2026-07-22 07:38:21.303263	2026-07-22 07:38:21.303263
c8b11f92-841d-4ed7-9e8f-16a7c0606581	6912df30-1a48-4569-af67-9885dccd1109	1	22	\N	6.000	3	18.95	\N	17.05	130.75	1	1	\N	\N	\N	\N	\N	2026-07-22 08:57:39.703606	2026-07-22 08:57:39.703606
0ba4bd05-2412-4778-bd68-442ebb37401e	6912df30-1a48-4569-af67-9885dccd1109	1	22	\N	6.000	3	18.95	\N	17.05	130.75	1	1	\N	\N	\N	\N	\N	2026-07-22 08:57:44.055467	2026-07-22 08:57:44.055467
2fdef10f-d454-49b0-ba05-02253261eb94	6912df30-1a48-4569-af67-9885dccd1109	1	22	\N	6.000	3	18.95	\N	17.05	130.75	1	1	\N	\N	\N	\N	\N	2026-07-22 08:58:06.831329	2026-07-22 08:58:06.831329
1f2f39cf-c640-47c5-ad19-76cfb7336af0	f3bc32e2-0d76-4bb6-9ed8-7f02b3021566	1	40	\N	1.000	4	48.50	\N	7.27	55.77	1	1	\N	\N	\N	\N	\N	2026-07-22 13:10:32.033338	2026-07-22 13:10:32.033338
f8dfb8fb-c9cf-4707-a25e-08891ec906cf	6912df30-1a48-4569-af67-9885dccd1109	1	3	\N	6.000	3	14.95	\N	13.45	103.15	1	1	\N	\N	\N	\N	\N	2026-07-22 08:59:37.303003	2026-07-22 08:59:37.303003
eba159c8-d3a7-4e37-b72f-f978f614a484	f3bc32e2-0d76-4bb6-9ed8-7f02b3021566	1	22	\N	6.000	3	18.95	\N	17.05	130.75	1	1	\N	\N	\N	\N	\N	2026-07-22 13:10:40.413659	2026-07-22 13:10:40.413659
ff9d4628-de05-4235-8985-77a404bef917	c03314a4-0319-467c-81c7-32b20fe42417	1	22	\N	14.000	3	18.95	\N	17.05	282.35	1	1	\N	\N	\N	\N	\N	2026-07-22 09:46:59.613455	2026-07-22 09:52:05.523451
8ba86319-8da8-4e8d-8b22-32609742be78	c03314a4-0319-467c-81c7-32b20fe42417	1	35	\N	1.000	10	3.50	\N	0.53	4.03	1	1	\N	\N	\N	\N	\N	2026-07-22 09:52:54.880993	2026-07-22 09:52:54.880993
d48cfb90-c985-49a1-852a-fd78c7a53856	161c21c8-61c7-4d8c-b4a0-3c9f7e760e3e	1	33	\N	1.000	10	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-07-22 10:00:56.259454	2026-07-22 10:00:56.259454
88c8f71f-7f2e-4d20-9a99-45763b5e9993	161c21c8-61c7-4d8c-b4a0-3c9f7e760e3e	1	26	\N	1.000	8	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-07-22 10:01:02.356456	2026-07-22 10:01:02.356456
051a4175-ccbf-4dc1-abe3-72fca6944978	6912df30-1a48-4569-af67-9885dccd1109	1	22	\N	7.000	3	18.95	\N	17.05	149.70	1	1	\N	\N	\N	\N	\N	2026-07-22 09:04:27.050279	2026-07-22 09:05:00.305674
8615ca1d-eff9-42af-94ca-28205b17a046	a0e2078d-2e1b-48a5-896f-ebc458458dc5	1	22	\N	6.000	3	18.95	\N	17.05	130.75	1	1	\N	\N	\N	\N	\N	2026-07-22 09:38:05.089701	2026-07-22 09:38:05.089701
b498d263-0032-4164-a86f-bd94dd3559e6	a0e2078d-2e1b-48a5-896f-ebc458458dc5	1	22	\N	6.000	3	18.95	\N	17.05	130.75	1	1	\N	\N	\N	\N	\N	2026-07-22 09:38:16.783026	2026-07-22 09:38:16.783026
dbc4b520-3d1b-4882-b191-3a9e77615ee5	a0e2078d-2e1b-48a5-896f-ebc458458dc5	1	40	\N	6.000	4	48.50	\N	43.65	334.65	1	1	\N	\N	\N	\N	\N	2026-07-22 09:39:01.261929	2026-07-22 09:39:01.261929
2f44bd6b-103a-48a3-bf11-4df77db0dee4	6bbea337-0a12-45a9-afb8-0e8eaeeaf9da	1	33	\N	1.000	10	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-07-22 10:02:48.07015	2026-07-22 10:02:48.07015
59036a04-20f8-46d3-a6d8-6b3f42ef1db5	55ac535e-f21a-4892-b7c4-ec5f824d8976	1	38	\N	12.000	2	35.95	\N	32.36	463.76	1	1	\N	\N	\N	\N	\N	2026-07-22 12:50:20.786962	2026-07-22 12:50:27.921625
5104c422-95a2-47dd-a479-4c4495ad8429	f3bc32e2-0d76-4bb6-9ed8-7f02b3021566	1	8	\N	10.000	1	18.00	\N	27.00	207.00	1	1	\N	\N	\N	\N	\N	2026-07-22 13:10:51.259823	2026-07-22 13:10:51.259823
c3a58ca7-53d8-486d-82f1-8c0c11abb70e	e52c169d-1614-411f-8bac-ab2db8c3a3af	1	22	\N	1.000	3	18.95	\N	2.84	21.79	1	1	\N	\N	\N	\N	\N	2026-07-22 13:15:06.136563	2026-07-22 13:15:06.136563
93bcc5ad-f691-4be6-b5b0-b722c580c8b7	6f23c0ff-e3eb-4ba6-9996-2225ed1afb14	1	8	\N	10.000	1	18.00	\N	27.00	207.00	1	1	\N	\N	\N	\N	\N	2026-07-22 13:18:38.19495	2026-07-22 13:18:38.19495
a27ba541-f423-4d5a-93f2-2e3d307409c4	4985babb-0bfa-4532-b61e-63a989fa9ee5	1	33	\N	1.000	10	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-07-23 06:17:57.236636	2026-07-23 06:17:57.236636
6ccae2c6-a2bf-41ff-9d56-1271470d0bba	4985babb-0bfa-4532-b61e-63a989fa9ee5	1	35	\N	1.000	10	3.50	\N	0.53	4.03	1	1	\N	\N	\N	\N	\N	2026-07-23 06:17:58.384805	2026-07-23 06:17:58.384805
0fbc5a5d-7728-4642-8be5-d657506d28bf	4985babb-0bfa-4532-b61e-63a989fa9ee5	1	6	\N	1.000	10	12.95	\N	1.94	14.89	1	1	\N	\N	\N	\N	\N	2026-07-23 06:18:07.103164	2026-07-23 06:18:07.103164
06c61e1b-e674-4797-934a-93c901c62647	4985babb-0bfa-4532-b61e-63a989fa9ee5	1	23	\N	1.000	3	19.95	\N	2.99	22.94	1	1	\N	\N	\N	\N	\N	2026-07-23 06:18:14.621774	2026-07-23 06:18:14.621774
5ebb2ecb-606f-4458-b391-e9ccd53d3e9a	93837ac6-d0d5-4bb6-b6fe-aae9278f91c1	1	34	\N	1.000	11	24.95	\N	3.74	28.69	1	1	\N	\N	\N	\N	\N	2026-07-23 06:28:24.172918	2026-07-23 06:28:24.172918
e47ae34a-9b60-4ebc-bdc4-f19dd5d140a8	93837ac6-d0d5-4bb6-b6fe-aae9278f91c1	1	33	\N	1.000	10	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-07-23 06:28:26.314101	2026-07-23 06:28:26.314101
43bc4a8b-2aa2-4432-ace4-fdb906a1f63f	93837ac6-d0d5-4bb6-b6fe-aae9278f91c1	1	35	\N	1.000	10	3.50	\N	0.53	4.03	1	1	\N	\N	\N	\N	\N	2026-07-23 06:28:31.170155	2026-07-23 06:28:31.170155
f6d9db53-0e13-4024-ba32-d079f722b791	e3f11cd0-804c-4c79-95e0-7b910cfacb6d	1	33	\N	1.000	10	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-07-23 06:37:53.720382	2026-07-23 06:37:53.720382
0e323e3f-21e9-4fc0-b625-5556190d4f32	ca973e0c-c7d6-4a65-a74e-895089e0527a	1	26	\N	1.000	8	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-07-23 08:30:30.304452	2026-07-23 08:30:30.304452
ef05a316-fa3a-498b-ac6c-083b1414b692	ca973e0c-c7d6-4a65-a74e-895089e0527a	1	35	\N	1.000	10	3.50	\N	0.53	4.03	1	1	\N	\N	\N	\N	\N	2026-07-23 08:30:32.164919	2026-07-23 08:30:32.164919
eb811959-6e09-4741-9398-86404a301cd1	ca973e0c-c7d6-4a65-a74e-895089e0527a	1	27	\N	1.000	8	8.95	\N	1.34	10.29	1	1	\N	\N	\N	\N	\N	2026-07-23 08:30:35.233104	2026-07-23 08:30:35.233104
a3b847fc-12af-4040-bd7e-d35a37531ace	0b90118f-6987-47c4-bc66-3325bd219b63	1	33	\N	1.000	10	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-07-23 10:53:29.421754	2026-07-23 10:53:29.421754
1d7b408b-bca3-464f-8866-02a2e93a7e77	fab12c2e-010a-4765-8f17-6e5cdcb041c6	1	6	\N	1.000	10	12.95	\N	1.94	14.89	1	1	\N	\N	\N	\N	\N	2026-07-23 11:09:39.851181	2026-07-23 11:09:39.851181
63ae1aed-7c96-4409-b29d-001104df6efd	6a6fff8d-1324-4573-9966-fbf179c43933	1	34	\N	1.000	11	24.95	\N	3.74	28.69	1	1	\N	\N	\N	\N	\N	2026-07-23 11:12:10.753556	2026-07-23 11:12:10.753556
af946230-2e17-4290-9b1d-dd8075c524ae	d8c30107-0eb2-4855-bb6a-51cd53505b5d	1	15	\N	1.000	3	7.95	\N	1.19	9.14	1	1	\N	\N	\N	\N	\N	2026-07-23 11:24:41.473734	2026-07-23 11:24:41.473734
3eaa8ae7-6d09-4045-a113-8403fb27ddcd	c269c10a-80e9-41ee-a08b-9c5360c2e4d1	1	33	\N	1.000	10	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-07-23 11:42:38.83135	2026-07-23 11:42:38.83135
07d8059c-1613-4ef9-8735-429b8a8b7cfc	bfd00d5b-5d48-491b-ade1-b18de6afcb68	1	35	\N	1.000	10	3.50	\N	0.53	4.03	1	1	\N	\N	\N	\N	\N	2026-07-24 09:14:33.979239	2026-07-24 09:14:33.979239
59b40550-934a-49fe-a9c4-57d2cf763086	bfd00d5b-5d48-491b-ade1-b18de6afcb68	1	6	\N	1.000	10	12.95	\N	1.94	14.89	1	1	\N	\N	\N	\N	\N	2026-07-24 09:14:36.067401	2026-07-24 09:14:36.067401
91f0d204-525b-4429-87e0-19da7cce0726	bfd00d5b-5d48-491b-ade1-b18de6afcb68	1	23	\N	1.000	3	19.95	\N	2.99	22.94	1	1	\N	\N	\N	\N	\N	2026-07-24 09:14:36.930097	2026-07-24 09:14:36.930097
2e542503-ef56-427d-aa56-f251a5f16a1d	bfd00d5b-5d48-491b-ade1-b18de6afcb68	1	33	\N	2.000	10	4.50	\N	0.67	9.67	1	1	\N	\N	\N	\N	\N	2026-07-24 09:14:32.778475	2026-07-24 09:14:40.136759
29c37980-ee0f-4194-8e73-2d172b48bc20	bfd00d5b-5d48-491b-ade1-b18de6afcb68	1	27	\N	2.000	8	8.95	\N	1.34	19.24	1	1	\N	\N	\N	\N	\N	2026-07-24 09:14:35.05233	2026-07-24 09:14:42.8027
52d787b4-66aa-4c7d-8165-e4eb490452c6	2eb60586-650e-4927-a018-945a436f01a3	1	34	\N	1.000	11	24.95	\N	3.74	28.69	1	1	\N	\N	\N	\N	\N	2026-07-27 10:47:21.203472	2026-07-27 10:47:21.203472
6e23b8b7-08eb-4b8d-9d44-ce19df5ca350	c62c4d60-b9e9-4c32-9c09-413b37bed03e	1	34	\N	1.000	11	24.95	\N	3.74	28.69	1	1	\N	\N	\N	\N	\N	2026-07-27 11:06:17.224757	2026-07-27 11:06:17.224757
23706447-b474-41aa-bf13-16b4c7cbdf26	c62c4d60-b9e9-4c32-9c09-413b37bed03e	1	26	\N	1.000	8	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-07-27 11:06:18.18815	2026-07-27 11:06:18.18815
91689055-9be3-4660-b8c3-4d391b883820	a2eebeb3-662c-46df-82ef-e6ad81058c5b	1	34	\N	1.000	11	24.95	\N	3.74	28.69	1	1	\N	\N	\N	\N	\N	2026-07-27 11:11:14.541821	2026-07-27 11:11:14.541821
b7ab0267-f870-4c63-a629-a9d8ff4ec446	14edb196-5851-4397-b37d-66322afe2a2d	1	17	\N	1.000	4	14.95	\N	2.24	17.19	1	1	\N	\N	\N	\N	\N	2026-07-27 11:41:21.011706	2026-07-27 11:41:21.011706
ead36a35-ca14-4d7a-a3ea-253a5c7f1c92	14edb196-5851-4397-b37d-66322afe2a2d	1	29	\N	1.000	2	3.00	\N	0.45	3.45	1	1	\N	\N	\N	\N	\N	2026-07-27 11:41:22.95286	2026-07-27 11:41:22.95286
dfd3846c-d448-4a28-a3b9-8ca448b87715	14edb196-5851-4397-b37d-66322afe2a2d	1	33	\N	1.000	10	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-07-27 11:41:26.678748	2026-07-27 11:41:26.678748
fb240bae-f8dd-44f6-a2c1-adec2e469b9e	14edb196-5851-4397-b37d-66322afe2a2d	1	24	\N	1.000	10	6.50	\N	0.97	7.47	1	1	\N	\N	\N	\N	\N	2026-07-27 11:41:31.089821	2026-07-27 11:41:31.089821
09ba3dde-72e4-4204-a20c-f6fc54501911	25f07dc1-0c1b-4fc4-a9a5-da5398c129a7	1	6	\N	1.000	10	12.95	\N	1.94	14.89	1	1	\N	\N	\N	\N	\N	2026-07-28 08:35:51.910496	2026-07-28 08:35:51.910496
53ccac63-2627-400b-9a06-a31defc275bc	334043fe-9a22-4e2d-b316-8b780d6717dd	1	35	\N	39.000	10	3.50	0.00	20.67	157.17	1	1	\N	\N	\N	\N	\N	2026-07-27 11:57:27.55425	2026-07-27 12:00:53.793565
2daa1cda-4176-4454-9962-bf7b395685dc	7a73337e-1c3b-4c14-9cd9-9e449851d7ee	1	33	\N	1.000	10	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-07-28 08:37:05.617411	2026-07-28 08:37:05.617411
75df28be-148d-4d58-8b7e-b23700c73df6	14edb196-5851-4397-b37d-66322afe2a2d	1	35	\N	8.000	10	3.50	0.00	4.24	32.24	1	1	\N	\N	\N	\N	\N	2026-07-27 11:41:38.718969	2026-07-27 11:41:40.704635
aa2727f8-3688-4418-8311-a37aca69a992	339d0029-bbdf-49c3-8fc1-cadba066f020	1	35	\N	8.000	10	3.50	0.00	4.24	32.24	1	1	\N	\N	\N	\N	\N	2026-07-27 12:05:43.063302	2026-07-27 12:05:57.423455
9d7531db-bb6a-48a0-ad2f-4dd88071cd62	334043fe-9a22-4e2d-b316-8b780d6717dd	1	33	\N	7.000	10	4.50	0.00	4.69	36.19	1	1	\N	\N	\N	\N	\N	2026-07-27 12:01:04.028001	2026-07-27 12:01:11.797804
1baefbda-8759-46a4-9e5a-481b968efa3e	e1dad131-4aa2-487a-b854-e002b43febe4	1	14	\N	6.000	3	7.95	0.00	7.14	54.84	1	1	\N	\N	\N	\N	\N	2026-07-27 11:51:57.848515	2026-07-27 11:52:04.996479
86610a16-6e1d-4927-a296-4dff3335a8b1	339d0029-bbdf-49c3-8fc1-cadba066f020	1	33	\N	1.000	10	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-07-27 12:06:00.883773	2026-07-27 12:06:00.883773
bb53ef57-3b0a-49ff-8c3a-7675beef36ec	334043fe-9a22-4e2d-b316-8b780d6717dd	1	29	\N	8.000	2	3.00	0.00	3.60	27.60	1	1	\N	\N	\N	\N	\N	2026-07-27 11:57:19.492979	2026-07-27 12:02:49.657691
457caf04-9066-449e-9207-dfd5032e25cc	339d0029-bbdf-49c3-8fc1-cadba066f020	1	34	\N	1.000	11	24.95	\N	3.74	28.69	1	1	\N	\N	\N	\N	\N	2026-07-27 12:06:25.596792	2026-07-27 12:06:25.596792
7aad72f1-dcda-4467-a4e4-b3291fa3774e	7a73337e-1c3b-4c14-9cd9-9e449851d7ee	1	35	\N	1.000	10	3.50	\N	0.53	4.03	1	1	\N	\N	\N	\N	\N	2026-07-28 08:37:06.113438	2026-07-28 08:37:06.113438
c97567bd-b8bb-4b18-80cb-d7e2a409b41a	334043fe-9a22-4e2d-b316-8b780d6717dd	1	34	\N	2.000	11	24.95	0.00	7.48	57.38	1	1	\N	\N	\N	\N	\N	2026-07-27 11:57:52.063543	2026-07-27 11:58:05.352011
7673bfbb-8c2d-4348-9ba6-8d5b05a3ae7c	e1dad131-4aa2-487a-b854-e002b43febe4	1	33	\N	43.000	10	4.50	0.00	28.81	222.31	1	1	\N	\N	\N	\N	\N	2026-07-27 11:53:52.219525	2026-07-27 11:54:14.627371
06d597bc-4b70-4daf-ad09-2c4a5a6bd851	e1dad131-4aa2-487a-b854-e002b43febe4	1	24	\N	1.000	10	6.50	\N	0.97	7.47	1	1	\N	\N	\N	\N	\N	2026-07-27 11:54:17.363299	2026-07-27 11:54:17.363299
0457e4dd-13a8-4616-976c-3690cf360ab0	e1dad131-4aa2-487a-b854-e002b43febe4	1	24	\N	1.000	10	6.50	\N	0.97	7.47	1	1	\N	\N	\N	\N	\N	2026-07-27 11:54:17.364518	2026-07-27 11:54:17.364518
3bdc832c-84d3-4922-b2f4-89433cf6bfb8	e1dad131-4aa2-487a-b854-e002b43febe4	1	24	\N	1.000	10	6.50	\N	0.97	7.47	1	1	\N	\N	\N	\N	\N	2026-07-27 11:54:17.366326	2026-07-27 11:54:17.366326
45dfea4a-0185-49da-b192-49b15e61b394	e1dad131-4aa2-487a-b854-e002b43febe4	1	29	\N	1.000	2	3.00	\N	0.45	3.45	1	1	\N	\N	\N	\N	\N	2026-07-27 11:54:21.872968	2026-07-27 11:54:21.872968
6271e9f4-bbae-4179-b0bf-ad7b9aafcd58	e1dad131-4aa2-487a-b854-e002b43febe4	1	29	\N	1.000	2	3.00	\N	0.45	3.45	1	1	\N	\N	\N	\N	\N	2026-07-27 11:54:21.874958	2026-07-27 11:54:21.874958
6c7731bb-8f8e-41b2-822f-02253513675e	e1dad131-4aa2-487a-b854-e002b43febe4	1	29	\N	1.000	2	3.00	\N	0.45	3.45	1	1	\N	\N	\N	\N	\N	2026-07-27 11:54:21.874878	2026-07-27 11:54:21.874878
67485443-5625-4c7e-a94b-5f2bd12b58df	339d0029-bbdf-49c3-8fc1-cadba066f020	1	14	\N	5.000	3	7.95	0.00	5.95	45.70	1	1	\N	\N	\N	\N	\N	2026-07-27 12:05:24.020325	2026-07-27 12:05:40.741901
ff619532-a8bf-4758-b35b-1f8b2e6b8240	334043fe-9a22-4e2d-b316-8b780d6717dd	1	17	\N	3.000	4	14.95	0.00	6.72	51.57	1	1	\N	\N	\N	\N	\N	2026-07-27 11:58:37.672211	2026-07-27 11:58:40.796651
772f1f7c-6379-4e5a-aa02-1e03202f15ce	7a73337e-1c3b-4c14-9cd9-9e449851d7ee	1	7	\N	1.000	10	15.50	\N	2.32	17.82	1	1	\N	\N	\N	\N	\N	2026-07-28 08:37:08.449407	2026-07-28 08:37:08.449407
d8b4694b-d8d9-41ef-b3b6-6417c36a8c58	334043fe-9a22-4e2d-b316-8b780d6717dd	1	14	\N	55.000	3	7.95	0.00	65.45	502.70	1	1	\N	\N	\N	\N	\N	2026-07-27 11:57:32.19648	2026-07-27 11:58:50.996358
641aaf21-eb96-447b-957c-6701f8f6273f	7a73337e-1c3b-4c14-9cd9-9e449851d7ee	1	22	\N	1.000	3	18.95	\N	2.84	21.79	1	1	\N	\N	\N	\N	\N	2026-07-28 08:37:10.534873	2026-07-28 08:37:10.534873
bb08b282-4f14-4b9f-b394-a8f9da604ab9	339d0029-bbdf-49c3-8fc1-cadba066f020	1	26	\N	5.000	8	4.50	0.00	3.35	25.85	1	1	\N	\N	\N	\N	\N	2026-07-27 12:06:27.286229	2026-07-27 12:06:39.850745
18ba8c39-0414-4479-b567-83246926603a	1e4579b0-f3ac-4064-ad89-169eb05ffa12	1	33	\N	1.000	10	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-07-27 12:40:42.285026	2026-07-27 12:40:42.285026
0fa40412-ee3b-4c69-a64f-7d7ff380d46e	62161eda-60af-4495-a576-c6d352e0b040	1	22	\N	6.000	3	18.95	\N	17.05	130.75	1	1	\N	\N	\N	\N	\N	2026-07-30 08:02:45.658538	2026-07-30 08:02:45.658538
ad1a2079-348d-444d-9521-5e2c8b994b15	25f07dc1-0c1b-4fc4-a9a5-da5398c129a7	1	33	\N	1.000	10	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-07-28 08:35:46.084042	2026-07-28 08:35:46.084042
eceb6819-6033-47e4-b91f-bad1ec20cae7	25f07dc1-0c1b-4fc4-a9a5-da5398c129a7	1	27	\N	1.000	8	8.95	\N	1.34	10.29	1	1	\N	\N	\N	\N	\N	2026-07-28 08:35:47.736182	2026-07-28 08:35:47.736182
c8d261c2-95b6-4953-ac94-12b890157c57	25f07dc1-0c1b-4fc4-a9a5-da5398c129a7	1	35	\N	1.000	10	3.50	\N	0.53	4.03	1	1	\N	\N	\N	\N	\N	2026-07-28 08:35:48.862868	2026-07-28 08:35:48.862868
a9fc9973-1df3-4710-ac3e-13f2ad711a4e	252baaaf-5ae9-4d3a-84fe-3e0878ca2b3f	1	34	\N	1.000	11	24.95	\N	3.74	28.69	1	1	\N	\N	\N	\N	\N	2026-07-30 08:09:34.433043	2026-07-30 08:09:34.433043
2d98704d-9f6b-45ad-81fd-5bece0c72ba8	252baaaf-5ae9-4d3a-84fe-3e0878ca2b3f	1	26	\N	1.000	8	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-07-30 08:09:36.701344	2026-07-30 08:09:36.701344
4fada699-b6aa-4fab-a3a8-46f5959d1340	05664a1e-fd9e-4045-81b4-fd5cb0d4110c	1	33	\N	1.000	10	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-07-30 10:59:22.210962	2026-07-30 10:59:22.210962
009afd61-8d4d-4e8b-aedd-67578ad61ab1	05664a1e-fd9e-4045-81b4-fd5cb0d4110c	1	34	\N	3.000	11	24.95	0.00	11.22	86.07	1	1	\N	\N	\N	\N	\N	2026-07-30 10:59:16.346175	2026-07-30 10:59:19.152889
3219fb05-0e3d-407b-9ceb-28ae888377c7	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	1	33	\N	1.000	10	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-07-30 11:03:28.882213	2026-07-30 11:03:28.882213
24a54d80-cdde-42f5-b07f-8b54bcccaf17	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	1	34	\N	1.000	11	24.95	\N	3.74	28.69	1	1	\N	\N	\N	\N	\N	2026-07-30 11:03:29.228538	2026-07-30 11:03:29.228538
2677919c-6ff0-4f03-953b-1aed8c5b9ed6	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	1	35	\N	1.000	10	3.50	\N	0.53	4.03	1	1	\N	\N	\N	\N	\N	2026-07-30 11:03:30.094573	2026-07-30 11:03:30.094573
684dbca4-1b5b-442c-b04e-38ac9ec05a49	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	1	26	\N	1.000	8	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-07-30 11:03:30.969293	2026-07-30 11:03:30.969293
60bdb137-4449-478f-89f0-bfb87031de1d	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	1	6	\N	1.000	10	12.95	\N	1.94	14.89	1	1	\N	\N	\N	\N	\N	2026-07-30 11:03:31.782608	2026-07-30 11:03:31.782608
29590691-0693-4331-94ce-974194ba3909	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	1	7	\N	1.000	10	15.50	\N	2.32	17.82	1	1	\N	\N	\N	\N	\N	2026-07-30 11:03:34.356156	2026-07-30 11:03:34.356156
b0ae0d88-2c64-4144-a1a3-d6bc652ac750	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	1	23	\N	1.000	3	19.95	\N	2.99	22.94	1	1	\N	\N	\N	\N	\N	2026-07-30 11:03:34.754628	2026-07-30 11:03:34.754628
56e7b02f-000f-424a-ad5e-ccabbe3b9b01	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	1	22	\N	1.000	3	18.95	\N	2.84	21.79	1	1	\N	\N	\N	\N	\N	2026-07-30 11:03:35.812122	2026-07-30 11:03:35.812122
31ab3eea-ee03-41b3-a3c4-85371dc0bafb	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	1	40	\N	1.000	4	48.50	\N	7.27	55.77	1	1	\N	\N	\N	\N	\N	2026-07-30 11:03:38.267279	2026-07-30 11:03:38.267279
b8be2734-fcda-4064-9c61-95203639bff1	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	1	41	\N	1.000	11	12.95	\N	1.94	14.89	1	1	\N	\N	\N	\N	\N	2026-07-30 11:03:40.198939	2026-07-30 11:03:40.198939
618b396d-a3c6-4f0d-8ff4-71356e4d8bf6	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	1	8	\N	1.000	1	18.00	\N	2.70	20.70	1	1	\N	\N	\N	\N	\N	2026-07-30 11:03:41.131779	2026-07-30 11:03:41.131779
e4e6a576-077b-439e-b839-0d376efc9dcd	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	1	2	\N	1.000	3	8.50	\N	1.27	9.78	1	1	\N	\N	\N	\N	\N	2026-07-30 11:03:42.386116	2026-07-30 11:03:42.386116
04e82bad-cda8-4d8a-b86b-e10dafd0878e	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	1	1	\N	1.000	3	8.50	\N	1.27	9.78	1	1	\N	\N	\N	\N	\N	2026-07-30 11:03:44.825146	2026-07-30 11:03:44.825146
b5803832-4454-4a64-a0fa-3ee04501fdf8	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	1	32	\N	1.000	2	22.00	\N	3.30	25.30	1	1	\N	\N	\N	\N	\N	2026-07-30 11:03:46.293039	2026-07-30 11:03:46.293039
386ae820-5140-4c64-980b-19b5f031881a	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	1	3	\N	1.000	3	14.95	\N	2.24	17.19	1	1	\N	\N	\N	\N	\N	2026-07-30 11:03:46.888748	2026-07-30 11:03:46.888748
e71c38c5-79a2-4da9-a6af-53ff9c9a68a0	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	1	30	\N	1.000	2	12.95	\N	1.94	14.89	1	1	\N	\N	\N	\N	\N	2026-07-30 11:03:47.869659	2026-07-30 11:03:47.869659
df91624a-1fbf-461b-8b4f-1843005c91e7	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	1	31	\N	1.000	10	9.50	\N	1.43	10.93	1	1	\N	\N	\N	\N	\N	2026-07-30 11:03:48.942467	2026-07-30 11:03:48.942467
9d50e53e-63a9-43a4-b549-33a9b45cea2a	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	1	14	\N	1.000	3	7.95	\N	1.19	9.14	1	1	\N	\N	\N	\N	\N	2026-07-30 11:03:50.477715	2026-07-30 11:03:50.477715
2508da59-e038-48f5-83a4-ee367a40cc5b	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	1	15	\N	1.000	3	7.95	\N	1.19	9.14	1	1	\N	\N	\N	\N	\N	2026-07-30 11:03:51.873497	2026-07-30 11:03:51.873497
0926fe43-6c16-49e9-9f23-ea997eabc162	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	1	38	\N	1.000	2	35.95	\N	5.39	41.34	1	1	\N	\N	\N	\N	\N	2026-07-30 11:03:52.91841	2026-07-30 11:03:52.91841
636240d4-20d7-41c8-9d88-483c0c8fd7dd	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	1	39	\N	1.000	3	42.00	\N	6.30	48.30	1	1	\N	\N	\N	\N	\N	2026-07-30 11:03:53.772073	2026-07-30 11:03:53.772073
58a70bbc-0cd8-4235-a863-fbda995a4fd6	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	1	36	\N	1.000	11	8.95	\N	1.34	10.29	1	1	\N	\N	\N	\N	\N	2026-07-30 11:03:55.573901	2026-07-30 11:03:55.573901
cfe138ee-8118-4095-930c-62848692fb6b	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	1	37	\N	1.000	2	39.95	\N	5.99	45.94	1	1	\N	\N	\N	\N	\N	2026-07-30 11:03:58.674644	2026-07-30 11:03:58.674644
5f865eee-5cee-441e-8f5a-1c7bef61d191	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	1	25	\N	1.000	10	6.50	\N	0.97	7.47	1	1	\N	\N	\N	\N	\N	2026-07-30 11:04:00.092868	2026-07-30 11:04:00.092868
a643b46e-be82-494c-97ae-2a97a02d7300	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	1	20	\N	1.000	2	45.00	\N	6.75	51.75	1	1	\N	\N	\N	\N	\N	2026-07-30 11:04:00.961324	2026-07-30 11:04:00.961324
f94064f7-5d81-4a82-8bf2-396bd15c52e9	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	1	21	\N	1.000	2	85.00	\N	12.75	97.75	1	1	\N	\N	\N	\N	\N	2026-07-30 11:04:04.812029	2026-07-30 11:04:04.812029
0fb6679e-9e53-482d-b60f-e17e57bf689e	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	1	11	\N	1.000	7	6.50	\N	0.97	7.47	1	1	\N	\N	\N	\N	\N	2026-07-30 11:04:05.646188	2026-07-30 11:04:05.646188
abc656b1-429c-43cd-bdcc-e314beea9ab7	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	1	9	\N	1.000	8	2.00	\N	0.30	2.30	1	1	\N	\N	\N	\N	\N	2026-07-30 11:04:09.489956	2026-07-30 11:04:09.489956
56bee23b-502b-4dfb-a921-be46058001f8	aed43720-0c89-4be2-bec4-cbba48662f56	1	33	\N	1.000	10	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:09.403741	2026-07-30 11:08:09.403741
e03e759a-a508-4893-94db-b03887785735	aed43720-0c89-4be2-bec4-cbba48662f56	1	34	\N	1.000	11	24.95	\N	3.74	28.69	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:10.203006	2026-07-30 11:08:10.203006
dd3707d4-0734-40a1-ae8f-a60710e95390	aed43720-0c89-4be2-bec4-cbba48662f56	1	26	\N	1.000	8	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:11.481972	2026-07-30 11:08:11.481972
caee2393-b3c7-47e3-ad50-97cf7b0bb3dd	aed43720-0c89-4be2-bec4-cbba48662f56	1	35	\N	1.000	10	3.50	\N	0.53	4.03	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:12.545959	2026-07-30 11:08:12.545959
2ebfd053-1673-45d1-be6e-1ad57ed2850b	aed43720-0c89-4be2-bec4-cbba48662f56	1	27	\N	1.000	8	8.95	\N	1.34	10.29	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:13.404353	2026-07-30 11:08:13.404353
7171ee74-f87c-48b7-8eef-5d4c2fd60f57	aed43720-0c89-4be2-bec4-cbba48662f56	1	6	\N	1.000	10	12.95	\N	1.94	14.89	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:14.228969	2026-07-30 11:08:14.228969
eb335d5f-d315-494b-8ba8-6e71c4ac526e	aed43720-0c89-4be2-bec4-cbba48662f56	1	23	\N	1.000	3	19.95	\N	2.99	22.94	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:15.320036	2026-07-30 11:08:15.320036
7d90a5bb-0e18-4955-a46f-fecc0dcabed2	aed43720-0c89-4be2-bec4-cbba48662f56	1	7	\N	1.000	10	15.50	\N	2.32	17.82	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:15.865839	2026-07-30 11:08:15.865839
67d4d718-89e0-4aa0-8abb-fb18da5b0c3e	aed43720-0c89-4be2-bec4-cbba48662f56	1	22	\N	1.000	3	18.95	\N	2.84	21.79	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:17.450073	2026-07-30 11:08:17.450073
9bb24066-6207-4bfd-b141-bcd64ac96809	aed43720-0c89-4be2-bec4-cbba48662f56	1	40	\N	1.000	4	48.50	\N	7.27	55.77	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:20.209673	2026-07-30 11:08:20.209673
a4431ca3-e4f1-4adb-96c5-2cf085c00ce2	aed43720-0c89-4be2-bec4-cbba48662f56	1	8	\N	1.000	1	18.00	\N	2.70	20.70	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:21.36	2026-07-30 11:08:21.36
6378c103-50fb-4c92-ad3a-e5d6e143835c	aed43720-0c89-4be2-bec4-cbba48662f56	1	41	\N	1.000	11	12.95	\N	1.94	14.89	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:22.410722	2026-07-30 11:08:22.410722
ca1034b7-7a74-432e-aa0d-5556edeb6bd7	aed43720-0c89-4be2-bec4-cbba48662f56	1	1	\N	1.000	3	8.50	\N	1.27	9.78	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:25.731086	2026-07-30 11:08:25.731086
0b031826-ed4b-4b53-92d6-6f6d7203386e	aed43720-0c89-4be2-bec4-cbba48662f56	1	2	\N	1.000	3	8.50	\N	1.27	9.78	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:26.919904	2026-07-30 11:08:26.919904
1eb98d3d-425b-4795-9eb8-4d790f7e19ab	aed43720-0c89-4be2-bec4-cbba48662f56	1	32	\N	1.000	2	22.00	\N	3.30	25.30	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:28.374796	2026-07-30 11:08:28.374796
fb2a5ff2-73de-4b13-8bce-8c3c07a6d165	aed43720-0c89-4be2-bec4-cbba48662f56	1	3	\N	1.000	3	14.95	\N	2.24	17.19	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:29.150091	2026-07-30 11:08:29.150091
d6522968-aaab-4ac9-a2c7-ce40965a6eb1	aed43720-0c89-4be2-bec4-cbba48662f56	1	30	\N	1.000	2	12.95	\N	1.94	14.89	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:30.323321	2026-07-30 11:08:30.323321
0b302ae2-3686-4171-af6a-aee071385c1b	aed43720-0c89-4be2-bec4-cbba48662f56	1	31	\N	1.000	10	9.50	\N	1.43	10.93	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:31.012167	2026-07-30 11:08:31.012167
869f0eff-a059-4934-8dca-df2fc5b117a7	aed43720-0c89-4be2-bec4-cbba48662f56	1	14	\N	1.000	3	7.95	\N	1.19	9.14	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:32.999856	2026-07-30 11:08:32.999856
0d8afc93-5148-44a3-b5ce-c47eda7a8e23	aed43720-0c89-4be2-bec4-cbba48662f56	1	15	\N	1.000	3	7.95	\N	1.19	9.14	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:33.544042	2026-07-30 11:08:33.544042
5f5b0a95-47a9-4515-b8a6-b75d121d19a1	aed43720-0c89-4be2-bec4-cbba48662f56	1	39	\N	1.000	3	42.00	\N	6.30	48.30	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:34.201363	2026-07-30 11:08:34.201363
8eb61fed-088c-4497-a723-8483f6b51612	aed43720-0c89-4be2-bec4-cbba48662f56	1	38	\N	1.000	2	35.95	\N	5.39	41.34	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:34.966337	2026-07-30 11:08:34.966337
a3c131ae-f2ce-483b-993a-8397e4ec4f04	aed43720-0c89-4be2-bec4-cbba48662f56	1	36	\N	1.000	11	8.95	\N	1.34	10.29	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:36.591556	2026-07-30 11:08:36.591556
c4ebf4af-5bf5-4912-b53e-4480ab2801c8	aed43720-0c89-4be2-bec4-cbba48662f56	1	25	\N	1.000	10	6.50	\N	0.97	7.47	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:37.801435	2026-07-30 11:08:37.801435
9a5ef48f-9ff1-4e04-947a-9085144aa56a	aed43720-0c89-4be2-bec4-cbba48662f56	1	24	\N	1.000	10	6.50	\N	0.97	7.47	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:38.836003	2026-07-30 11:08:38.836003
2ee32214-c53b-450a-8f71-259ff8e8b64e	aed43720-0c89-4be2-bec4-cbba48662f56	1	20	\N	1.000	2	45.00	\N	6.75	51.75	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:39.90961	2026-07-30 11:08:39.90961
66f4487e-568f-44a0-b1e8-3577c3083793	aed43720-0c89-4be2-bec4-cbba48662f56	1	11	\N	1.000	7	6.50	\N	0.97	7.47	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:41.26356	2026-07-30 11:08:41.26356
8a4aa109-0eb5-4f4f-a947-2a3af2572bcf	aed43720-0c89-4be2-bec4-cbba48662f56	1	29	\N	1.000	2	3.00	\N	0.45	3.45	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:42.879429	2026-07-30 11:08:42.879429
57b59fb5-283c-4ef7-819d-802636453606	aed43720-0c89-4be2-bec4-cbba48662f56	1	10	\N	1.000	8	2.00	\N	0.30	2.30	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:43.839824	2026-07-30 11:08:43.839824
d22ec36c-2149-4182-9101-041698abc9b0	aed43720-0c89-4be2-bec4-cbba48662f56	1	17	\N	1.000	4	14.95	\N	2.24	17.19	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:46.279334	2026-07-30 11:08:46.279334
f47ea984-a5a3-48f2-b7f3-ab7fac35a087	aed43720-0c89-4be2-bec4-cbba48662f56	1	19	\N	1.000	10	32.00	\N	4.80	36.80	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:47.55778	2026-07-30 11:08:47.55778
ff30efb2-b796-4940-942e-06124c475d0b	aed43720-0c89-4be2-bec4-cbba48662f56	1	16	\N	1.000	4	12.50	\N	1.88	14.38	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:49.131627	2026-07-30 11:08:49.131627
483ad10c-cc64-4152-9444-adcf958d6bf8	aed43720-0c89-4be2-bec4-cbba48662f56	1	13	\N	1.000	7	1.50	\N	0.22	1.73	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:49.992667	2026-07-30 11:08:49.992667
25885d6f-8d83-4f16-bdde-d79d9d562687	aed43720-0c89-4be2-bec4-cbba48662f56	1	5	\N	1.000	10	3.50	\N	0.53	4.03	1	1	\N	\N	\N	\N	\N	2026-07-30 11:08:51.554629	2026-07-30 11:08:51.554629
3d489c03-53e5-4959-9aa4-ca4623298d00	a6a8a674-e57e-4809-9e86-b7d517ef7d4f	1	5	\N	1.000	10	3.50	\N	0.53	4.03	1	1	\N	\N	\N	\N	\N	2026-07-31 07:55:13.924815	2026-07-31 07:55:13.924815
69e286a0-c72f-4eed-b1b0-13e36b0738fb	a6a8a674-e57e-4809-9e86-b7d517ef7d4f	1	5	\N	1.000	10	3.50	\N	0.53	4.03	1	1	\N	\N	\N	\N	\N	2026-07-31 07:55:26.733702	2026-07-31 07:55:26.733702
3740c595-647d-45c9-ab09-c6df732527bd	a6a8a674-e57e-4809-9e86-b7d517ef7d4f	1	8	\N	1.000	1	18.00	\N	2.70	20.70	1	1	\N	\N	\N	\N	\N	2026-07-31 08:10:55.908095	2026-07-31 08:10:55.908095
d7066b4a-028f-4168-bed2-ba481b330984	a6a8a674-e57e-4809-9e86-b7d517ef7d4f	1	5	\N	1.000	10	3.50	\N	0.53	4.03	1	1	\N	\N	\N	\N	\N	2026-07-31 08:11:13.306412	2026-07-31 08:11:13.306412
1031cc3b-ba9e-4a37-98b5-47eb31584373	a6a8a674-e57e-4809-9e86-b7d517ef7d4f	1	7	\N	1.000	10	15.50	\N	2.32	17.82	1	1	\N	\N	\N	\N	\N	2026-07-31 08:11:15.756361	2026-07-31 08:11:15.756361
b7430534-5772-4d32-b61a-03d6a852a7d6	a6a8a674-e57e-4809-9e86-b7d517ef7d4f	1	5	\N	1.000	10	3.50	\N	0.53	4.03	1	1	\N	\N	\N	\N	\N	2026-07-31 08:13:46.495252	2026-07-31 08:13:46.495252
bf2cc057-14b0-48b4-87b1-3778ac6492f8	a6a8a674-e57e-4809-9e86-b7d517ef7d4f	1	5	\N	1.000	10	3.50	\N	0.53	4.03	1	1	\N	\N	\N	\N	\N	2026-07-31 08:14:02.39028	2026-07-31 08:14:02.39028
27470c01-36b2-4834-8c19-93382762effb	a6a8a674-e57e-4809-9e86-b7d517ef7d4f	1	5	\N	1.000	10	3.50	\N	0.53	4.03	1	1	\N	\N	\N	\N	\N	2026-07-31 08:15:35.274217	2026-07-31 08:15:35.274217
f4faad8b-4385-4571-8183-65f463e8fc8e	a6a8a674-e57e-4809-9e86-b7d517ef7d4f	1	5	\N	1.000	10	3.50	\N	0.53	4.03	1	1	\N	\N	\N	\N	\N	2026-07-31 09:02:33.183981	2026-07-31 09:02:33.183981
28eec3b9-c269-492c-8cb2-7b972a193f07	d397fad3-1a15-4db2-a6db-bec5bf2b9297	1	5	\N	1.000	10	3.50	\N	0.53	4.03	1	1	\N	\N	\N	\N	\N	2026-07-31 09:05:01.38503	2026-07-31 09:05:01.38503
4c75cad2-32ed-404b-bdb2-0fea7d01d1e9	8f1bca0d-e6d0-434e-9895-8a68343bd13f	1	34	\N	1.000	11	24.95	\N	3.74	28.69	1	1	\N	\N	\N	\N	\N	2026-07-31 09:06:32.413296	2026-07-31 09:06:32.413296
8766bb34-4181-41eb-b982-fbbcf88e9eea	8f1bca0d-e6d0-434e-9895-8a68343bd13f	1	35	\N	1.000	10	3.50	\N	0.53	4.03	1	1	\N	\N	\N	\N	\N	2026-07-31 09:06:37.665634	2026-07-31 09:06:37.665634
2a6bc8ad-effd-40ff-8dd1-a53cfe03af04	8f1bca0d-e6d0-434e-9895-8a68343bd13f	1	26	\N	1.000	8	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-07-31 09:09:34.502105	2026-07-31 09:09:34.502105
d1188908-ad54-435f-a44f-34edc88a71b3	deb0a121-45eb-439b-ab8b-b6e4979a960f	1	1	\N	1.000	3	8.50	\N	1.27	9.78	1	1	\N	\N	\N	\N	\N	2026-07-31 10:23:28.811259	2026-07-31 10:23:28.811259
c74b174f-03ff-4328-a2d1-2499c6db8116	8f1bca0d-e6d0-434e-9895-8a68343bd13f	1	27	\N	1.000	8	8.95	\N	1.34	10.29	1	1	\N	\N	\N	\N	\N	2026-07-31 09:09:54.694958	2026-07-31 09:09:54.694958
ad678553-94ab-4971-a378-79cdc7407b5e	8f1bca0d-e6d0-434e-9895-8a68343bd13f	1	33	\N	3.000	10	4.50	0.00	2.01	15.51	1	1	\N	\N	\N	\N	\N	2026-07-31 09:06:29.74906	2026-07-31 09:10:06.196178
b4df4a8b-2aab-4e1e-931a-9bb1f6836120	8f1bca0d-e6d0-434e-9895-8a68343bd13f	1	2	\N	1.000	3	8.50	\N	1.27	9.78	1	1	\N	\N	\N	\N	\N	2026-07-31 09:10:16.705865	2026-07-31 09:10:16.705865
16dda717-453f-4809-988b-9fb138713b7e	3469e4f7-1aef-475b-951b-2cd3ec962029	1	33	\N	1.000	10	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-07-31 09:13:26.935613	2026-07-31 09:13:26.935613
3dac596a-ee7d-42bb-a60e-687232ad6914	3469e4f7-1aef-475b-951b-2cd3ec962029	1	34	\N	1.000	11	24.95	\N	3.74	28.69	1	1	\N	\N	\N	\N	\N	2026-07-31 09:13:28.828299	2026-07-31 09:13:28.828299
bfda65d3-942a-4ff2-a594-0fba10774323	3b93a016-e219-4b71-a2a8-9ef8886d2aae	1	35	\N	1.000	10	3.50	\N	0.53	4.03	1	1	\N	\N	\N	\N	\N	2026-07-31 10:04:03.77136	2026-07-31 10:04:03.77136
e3f3dec6-e97d-4ccc-a657-d5af8492a3c9	3b93a016-e219-4b71-a2a8-9ef8886d2aae	1	33	\N	2.000	10	4.50	0.00	1.34	10.34	1	1	\N	\N	\N	\N	\N	2026-07-31 10:04:01.683437	2026-07-31 10:04:08.682396
547c4444-4eba-4644-8c5f-c7b9ebb12d6f	3b93a016-e219-4b71-a2a8-9ef8886d2aae	1	26	\N	2.000	8	4.50	0.00	1.34	10.34	1	1	\N	\N	\N	\N	\N	2026-07-31 10:04:11.251885	2026-07-31 10:11:56.060163
d07c8dbf-9fa2-4a0a-8707-a2aa064ca4b0	3b93a016-e219-4b71-a2a8-9ef8886d2aae	1	3	\N	1.000	3	14.95	\N	2.24	17.19	1	1	\N	\N	\N	\N	\N	2026-07-31 10:15:45.224658	2026-07-31 10:15:45.224658
29a4b69c-6526-4190-a184-d3858bcff5a0	3b93a016-e219-4b71-a2a8-9ef8886d2aae	1	31	\N	1.000	10	9.50	\N	1.43	10.93	1	1	\N	\N	\N	\N	\N	2026-07-31 10:15:49.025826	2026-07-31 10:15:49.025826
75b288df-35b6-47b0-a7b0-b88eae65d53e	3b93a016-e219-4b71-a2a8-9ef8886d2aae	1	6	\N	2.000	10	12.95	0.00	3.88	29.78	1	1	\N	\N	\N	\N	\N	2026-07-31 10:11:57.812395	2026-07-31 10:17:52.20772
4b933232-8cf5-4f0e-bdbe-c71dc28a2abe	3b93a016-e219-4b71-a2a8-9ef8886d2aae	1	41	\N	1.000	11	12.95	\N	1.94	14.89	1	1	\N	\N	\N	\N	\N	2026-07-31 10:17:59.648682	2026-07-31 10:17:59.648682
8aba90af-f834-4d9e-914b-fa4f767208ee	3b93a016-e219-4b71-a2a8-9ef8886d2aae	1	2	\N	1.000	3	8.50	\N	1.27	9.78	1	1	\N	\N	\N	\N	\N	2026-07-31 10:18:05.215849	2026-07-31 10:18:05.215849
e1cbe09e-9d2d-4b43-ac94-b5474cc9ee5d	3b93a016-e219-4b71-a2a8-9ef8886d2aae	1	30	\N	2.000	2	12.95	0.00	3.88	29.78	1	1	\N	\N	\N	\N	\N	2026-07-31 10:15:52.745689	2026-07-31 10:18:10.809955
01a3a64f-e63c-4dba-814c-b1eb2fd572c9	3b93a016-e219-4b71-a2a8-9ef8886d2aae	1	37	\N	2.000	2	39.95	0.00	11.98	91.88	1	1	\N	\N	\N	\N	\N	2026-07-31 10:18:17.209663	2026-07-31 10:18:42.149971
985a3523-8ee1-4f95-8d60-a63ee3d0669b	c37191dd-8749-477b-bbe3-402c2e09a31d	1	35	\N	1.000	10	3.50	\N	0.53	4.03	1	1	\N	\N	\N	\N	\N	2026-08-01 06:17:48.115572	2026-08-01 06:17:48.115572
8a695e84-5f8d-449e-a465-2687c6231d69	c37191dd-8749-477b-bbe3-402c2e09a31d	1	22	\N	1.000	3	18.95	\N	2.84	21.79	1	1	\N	\N	\N	\N	\N	2026-08-01 06:17:53.789366	2026-08-01 06:17:53.789366
193123c7-f694-4109-8278-7152e8ee40e3	c37191dd-8749-477b-bbe3-402c2e09a31d	1	8	\N	1.000	1	18.00	\N	2.70	20.70	1	1	\N	\N	\N	\N	\N	2026-08-01 06:17:58.694945	2026-08-01 06:17:58.694945
38309620-dde9-40fb-afe7-8585c4d506e7	b627c8a2-bc74-4bb8-9b28-0790aab1266a	1	35	\N	4.000	10	3.50	0.00	2.12	16.12	1	1	\N	\N	\N	\N	\N	2026-08-01 06:23:27.018799	2026-08-01 06:26:36.490242
22424e4a-f946-4d60-95b0-b88b433bfaba	de7f9185-027d-4156-bf2c-7b03cd948e88	1	34	\N	1.000	11	24.95	\N	3.74	28.69	1	1	\N	\N	\N	\N	\N	2026-08-01 06:46:39.12072	2026-08-01 06:46:39.12072
68e41369-1e11-49be-a2e0-c00224d5249c	de7f9185-027d-4156-bf2c-7b03cd948e88	1	35	\N	1.000	10	3.50	\N	0.53	4.03	1	1	\N	\N	\N	\N	\N	2026-08-01 06:46:42.403777	2026-08-01 06:46:42.403777
7cdb72dd-dda4-4995-adb1-9b382914737f	7750e699-09d6-45b6-a904-e408f9dabaef	1	27	\N	1.000	8	8.95	\N	1.34	10.29	1	1	\N	\N	\N	\N	\N	2026-08-03 05:30:37.149505	2026-08-03 05:30:37.149505
e0522c54-1fa6-44a3-87e4-199768e2edf1	7750e699-09d6-45b6-a904-e408f9dabaef	1	34	\N	10.000	11	24.95	0.00	37.40	286.90	1	1	\N	\N	\N	\N	\N	2026-08-03 05:21:59.896905	2026-08-03 05:27:04.74232
1e2e2772-10ae-41ab-88da-aae7d0bdda8f	7750e699-09d6-45b6-a904-e408f9dabaef	1	23	\N	5.000	3	19.95	0.00	14.95	114.70	1	1	\N	\N	\N	\N	\N	2026-08-03 05:30:41.647005	2026-08-03 05:30:49.305968
1e9d2c64-42f1-486f-885e-80c24ace7d07	b014c45e-d107-4546-9852-0ceafe1d6282	1	6	\N	1.000	10	12.95	\N	1.94	14.89	1	1	\N	\N	\N	\N	\N	2026-08-03 09:37:23.839043	2026-08-03 09:37:23.839043
219a3577-e88b-40ac-90f3-69058c7bc3b4	b014c45e-d107-4546-9852-0ceafe1d6282	1	6	\N	1.000	10	12.95	\N	1.94	14.89	1	1	\N	\N	\N	\N	\N	2026-08-03 09:37:24.031032	2026-08-03 09:37:24.031032
8aa1edd8-b381-4001-86e3-a1fb155da3ee	b014c45e-d107-4546-9852-0ceafe1d6282	1	6	\N	1.000	10	12.95	\N	1.94	14.89	1	1	\N	\N	\N	\N	\N	2026-08-03 09:52:30.371728	2026-08-03 09:52:30.371728
e05871cc-8e38-47ea-a263-3ec6598063c3	b014c45e-d107-4546-9852-0ceafe1d6282	1	8	\N	1.000	1	18.00	\N	2.70	20.70	1	1	\N	\N	\N	\N	\N	2026-08-03 09:52:47.351	2026-08-03 09:52:47.351
6568ef16-8440-4ed5-a227-42b9b35e3379	35aa084d-d355-48f6-88d7-8817ee31408b	1	33	\N	1.000	10	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-08-03 10:34:16.375017	2026-08-03 10:34:16.375017
c023915d-001c-4f60-a953-4b3d75fd2c5d	dabc964f-45e9-4234-a23d-4e2fddcdaaee	1	33	\N	1.000	10	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-08-03 10:35:59.745194	2026-08-03 10:35:59.745194
46fc015d-9c56-41ae-a63d-008b8b5a121b	9f357046-da05-423c-8753-9c3031e83a44	1	35	\N	1.000	10	3.50	\N	0.53	4.03	1	1	\N	\N	\N	\N	\N	2026-08-03 10:36:22.869759	2026-08-03 10:36:22.869759
73cadd46-90a8-4bbc-9c8e-532860dce078	9f357046-da05-423c-8753-9c3031e83a44	1	33	\N	1.000	10	4.50	\N	0.67	5.17	1	1	\N	\N	\N	\N	\N	2026-08-03 10:36:40.970869	2026-08-03 10:36:40.970869
24533824-27f2-4f66-a0f5-3f7651445f2c	73419d52-4182-413f-a4da-14cc77155a65	1	8	\N	1.000	1	18.00	\N	2.70	20.70	1	1	\N	\N	\N	\N	\N	2026-08-03 11:21:55.824097	2026-08-03 11:21:55.824097
8b6504dc-44d0-470d-987b-86bc2531f5f9	4620d829-c1f1-4885-9686-3bf3796b3cbb	1	5	\N	1.000	10	3.50	\N	0.53	4.03	1	1	\N	\N	\N	\N	\N	2026-08-03 11:41:25.108536	2026-08-03 11:41:25.108536
d080e101-795e-4b91-83bf-54f42e093a51	dc47e996-a7df-481b-869d-6f85fb5f51c5	1	5	\N	1.000	10	3.50	\N	0.53	4.03	1	1	\N	\N	\N	\N	\N	2026-08-03 11:53:43.345966	2026-08-03 11:53:43.345966
e52e116d-eefe-49b5-9c0b-50348d1fddd3	a0f2df8d-1907-41dd-927d-749d3e09730b	1	35	\N	2.000	10	3.50	0.00	1.06	8.06	1	1	\N	\N	\N	\N	\N	2026-08-11 11:23:49.615303	2026-08-11 11:24:21.113664
f4428499-c010-4929-ace7-eccf3516aaea	ecaed928-ee99-437d-8e8d-e46d4597aff8	1	35	\N	1.000	10	3.50	\N	0.53	4.03	1	1	\N	\N	\N	\N	\N	2026-08-11 11:27:20.653529	2026-08-11 11:27:20.653529
815182dc-1568-44d2-bb2c-a117f9c93eab	ecaed928-ee99-437d-8e8d-e46d4597aff8	1	33	\N	2.000	10	4.50	0.00	1.34	10.34	1	1	\N	\N	\N	\N	\N	2026-08-11 11:27:24.234874	2026-08-11 11:27:27.782754
372de7a0-f4d8-493e-8390-01e637fe09c7	ecaed928-ee99-437d-8e8d-e46d4597aff8	1	7	\N	1.000	10	15.50	\N	2.32	17.82	1	1	\N	\N	\N	\N	\N	2026-08-11 11:42:08.773152	2026-08-11 11:42:08.773152
93f1b412-253e-4753-b4f0-ccfb511b89d6	ecaed928-ee99-437d-8e8d-e46d4597aff8	1	26	\N	3.000	8	4.50	0.00	2.01	15.51	1	1	\N	\N	\N	\N	\N	2026-08-11 11:32:42.457526	2026-08-11 11:43:14.637721
\.


--
-- Data for Name: carts; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.carts (id, cart_number, organization_id, store_id, customer_id, guest_identifier, guest_email, guest_phone, cart_status, cart_type, channel, payment_method, payment_gateway, device_info, created_by_user_id, cashier_id, pos_terminal_id, subtotal, discount_amount, tax_amount, shipping_amount, total_amount, coupon_code, discount_code, promotional_credits, shipping_address, billing_address, shipping_method, converted_to_order_id, converted_at, last_activity_at, expires_at, created_at, updated_at, metadata, notes) FROM stdin;
4985babb-0bfa-4532-b61e-63a989fa9ee5	CART-20260723-9518	1	1	10	guest-1784787475014	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	9	40.90	0.00	6.13	0.00	47.03			0.00	{}	{}	standard	bb618d80-6189-4df6-91a2-263755459002	2026-07-23 06:18:29.318977	2026-07-23 06:18:14.621774	2026-07-25 06:17:55.014	2026-07-23 06:17:54.658833	2026-07-23 06:18:29.318977	{}	
56324b30-f1ce-48c0-aea8-be446dff854c	CART-20260720-2812	1	1	1	guest-1784545376393			converted	standard	pos	\N	\N	{}	4	1	8	42.40	0.00	6.35	0.00	48.75			0.00	{}	{}	standard	f31d00e6-3197-4290-8783-5cb985915208	2026-07-20 11:03:02.714912	2026-07-20 11:02:59.969627	2026-07-22 11:02:56.393	2026-07-20 11:02:55.971192	2026-07-20 11:03:02.714912	{}	
b4ab7b6b-7e3c-43a9-b297-7e4950fd3b96	CART-20260720-1591	1	1	1	guest-1784533778552			converted	standard	pos	\N	\N	{}	4	1	8	223.45	0.00	33.50	0.00	256.95			0.00	{}	{}	standard	60f69ca5-9235-44c5-bfd6-acb162e8e2c3	2026-07-23 11:16:48.139645	2026-07-20 08:20:59.637092	2026-07-22 07:49:38.552	2026-07-20 07:49:38.125875	2026-07-23 11:16:48.139645	{}	
6912df30-1a48-4569-af67-9885dccd1109	CART-20260722-5190	1	1	1	guest-1784710652014	guest@gmail.com	03123456789	converted	standard	pos	\N	\N	{}	4	1	8	563.45	0.00	81.65	0.00	645.10			0.00	{}	{}	standard	97812e45-0a4d-437b-8641-cf865d76f464	2026-07-22 09:05:40.908668	2026-07-22 09:05:00.305674	2026-07-24 08:57:32.014	2026-07-22 08:57:32.675506	2026-07-22 09:05:40.908668	{}	
786d75ea-064f-4ec6-9f1e-15a9cd122adc	CART-20260720-8381	1	1	1	guest-1784535682760			converted	standard	pos	\N	\N	{}	4	1	8	188.51	0.00	28.27	0.00	216.78			0.00	{}	{}	standard	dd2f97c3-8864-4ea1-92d9-fba5fe3326bf	2026-07-20 08:21:31.507485	2026-07-20 08:21:29.199133	2026-07-22 08:21:22.76	2026-07-20 08:21:22.383376	2026-07-20 08:21:31.507485	{}	
93837ac6-d0d5-4bb6-b6fe-aae9278f91c1	CART-20260723-7419	1	1	10	guest-1784788102635	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	9	32.95	0.00	4.94	0.00	37.89			0.00	{}	{}	standard	fc451cb2-50b8-45d3-80a5-e478c795339a	2026-07-23 06:28:34.485405	2026-07-23 06:28:31.170155	2026-07-25 06:28:22.635	2026-07-23 06:28:22.132877	2026-07-23 06:28:34.485405	{}	
b6933e63-f839-4d53-a4da-501b113d878a	CART-20260720-5210	1	1	1	guest-1784547805925			converted	standard	pos	\N	\N	{}	4	1	8	28.45	0.00	4.27	0.00	32.72			0.00	{}	{}	standard	3b0666c1-9ea5-4698-9c8f-06dc39329f39	2026-07-20 11:43:31.225812	2026-07-20 11:43:28.681905	2026-07-22 11:43:25.925	2026-07-20 11:43:25.460484	2026-07-20 11:43:31.225812	{}	
1b665464-53d1-4f00-9076-ccc0b9d8d058	CART-20260720-2704	1	1	1	guest-1784536257086			converted	standard	pos	\N	\N	{}	4	1	8	326.45	0.00	48.96	0.00	375.41			0.00	{}	{}	standard	84b01ab7-92d9-42ba-9741-663c8ecc6b11	2026-07-20 08:31:16.39892	2026-07-20 08:31:13.054929	2026-07-22 08:30:57.086	2026-07-20 08:30:56.652387	2026-07-20 08:31:16.39892	{}	
e3f11cd0-804c-4c79-95e0-7b910cfacb6d	CART-20260723-1894	1	1	10	guest-1784788672194	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	9	4.50	0.00	0.67	0.00	5.17			0.00	{}	{}	standard	604348ff-bd9a-480f-9d9e-381151944683	2026-07-23 06:37:55.624729	2026-07-23 06:37:53.720382	2026-07-25 06:37:52.194	2026-07-23 06:37:51.708596	2026-07-23 06:37:55.624729	{}	
def65158-beb8-4300-83d5-fba82480ba3e	CART-20260722-3236	1	1	2	guest-1784705821210	jane@example.com	+15551234567	converted	standard	pos	\N	\N	{}	4	1	9	37.90	0.00	5.68	0.00	43.58			0.00	{}	{}	standard	7d954505-fbde-4d37-8ea3-f733e5f87429	2026-07-22 07:37:06.207043	2026-07-22 07:37:04.43384	2026-07-24 07:37:01.21	2026-07-22 07:37:01.361059	2026-07-22 07:37:06.207043	{}	
ca973e0c-c7d6-4a65-a74e-895089e0527a	CART-20260723-5671	1	1	10	guest-1784795426837	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	9	16.95	0.00	2.54	0.00	19.49			0.00	{}	{}	standard	1c15059e-0602-4345-8a9e-3d7aec6962a4	2026-07-23 08:31:04.30609	2026-07-23 08:30:35.233104	2026-07-25 08:30:26.837	2026-07-23 08:30:26.53071	2026-07-23 08:31:04.30609	{}	
156c1666-36f7-4a09-92e5-22dc40cbeba8	CART-20260722-2445	1	1	10	guest-1784714730141	Guest@gmail.com	12345678	active	standard	pos	\N	\N	{}	4	1	8	0.00	0.00	0.00	0.00	0.00			0.00	{}	{}	standard	\N	2026-07-22 12:49:41.669	2026-07-23 10:51:51.017684	2026-07-24 10:05:30.141	2026-07-22 10:05:30.320101	2026-07-23 10:51:51.019376	{}	
3dfb4094-071e-40a5-bda1-c6e2159e0c7c	CART-20260721-8604	1	1	2	guest-1784617884262	jane@example.com	+15551234567	converted	standard	pos	\N	\N	{}	4	1	9	33.95	0.00	5.08	0.00	39.03			0.00	{}	{}	standard	9bc4c20b-4fb5-4c28-9104-8552a3195649	2026-07-21 07:11:31.105103	2026-07-21 07:11:29.106134	2026-07-23 07:11:24.262	2026-07-21 07:11:24.443566	2026-07-21 07:11:31.105103	{}	
37447439-4024-451a-8807-b8cfca010487	CART-20260721-5998	1	1	1	guest-1784617959275	guest@gmail.com	03123456789	converted	standard	pos	\N	\N	{}	4	1	9	303.45	0.00	45.52	0.00	348.97			0.00	{}	{}	standard	4898af5e-ea60-4761-af8d-08aef13b9274	2026-07-21 07:12:48.86144	2026-07-21 07:12:46.320548	2026-07-23 07:12:39.275	2026-07-21 07:12:39.140278	2026-07-21 07:12:48.86144	{}	
a0e2078d-2e1b-48a5-896f-ebc458458dc5	CART-20260722-0176	1	1	10	guest-1784713045305	adilfarooq@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	8	518.40	0.00	77.75	0.00	596.15			0.00	{}	{}	standard	25f422ca-343b-4dac-b452-96f9e58dbc82	2026-07-23 11:07:35.557081	2026-07-22 09:39:01.261929	2026-07-24 09:37:25.305	2026-07-22 09:37:25.448659	2026-07-23 11:07:35.557081	{}	
d8c30107-0eb2-4855-bb6a-51cd53505b5d	CART-20260720-1376	1	1	1	guest-1784533000221			converted	standard	pos	\N	\N	{}	4	1	8	32.90	0.00	4.93	0.00	37.83			0.00	{}	{}	standard	28789b9c-918d-4697-ae9c-19feed267c18	2026-07-23 11:24:59.685238	2026-07-23 11:24:41.473734	2026-07-22 07:36:40.221	2026-07-20 07:36:39.818643	2026-07-23 11:24:59.685238	{}	
ea501529-ce47-4e5f-898a-6c9e2304c995	CART-20260721-3255	1	1	1	guest-1784617904837	guest@gmail.com	03123456789	converted	standard	pos	\N	\N	{}	4	1	9	42.40	0.00	6.35	0.00	48.75			0.00	{}	{}	standard	52a68651-bf78-476f-952a-b29109ba9b2f	2026-07-21 07:12:23.889965	2026-07-21 07:11:48.970178	2026-07-23 07:11:44.837	2026-07-21 07:11:44.934315	2026-07-21 07:12:23.889965	{}	
161c21c8-61c7-4d8c-b4a0-3c9f7e760e3e	CART-20260722-6841	1	1	2	guest-1784714453238	jane@example.com	+15551234567	converted	standard	pos	\N	\N	{}	4	1	8	9.00	0.00	1.34	0.00	10.34			0.00	{}	{}	standard	98ec425c-3a11-4f1e-ac90-5d96ab6e5819	2026-07-22 10:02:26.789285	2026-07-22 10:02:21.154038	2026-07-24 10:00:53.238	2026-07-22 10:00:53.392727	2026-07-22 10:02:26.789285	{}	
c03314a4-0319-467c-81c7-32b20fe42417	CART-20260722-7090	1	1	10	guest-1784713472673	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	8	268.80	0.00	17.58	0.00	286.38			0.00	{}	{}	standard	c2fc4fa8-a714-458d-9656-f82085a9c28b	2026-07-22 09:54:51.07443	2026-07-22 09:52:54.880993	2026-07-24 09:44:32.673	2026-07-22 09:44:32.790789	2026-07-22 09:54:51.07443	{}	
d8d59145-07e8-467c-b016-9555c1af7a88	CART-20260722-2983	1	1	2	guest-1784705847658	jane@example.com	+15551234567	converted	standard	pos	\N	\N	{}	4	1	9	291.00	0.00	43.65	0.00	334.65			0.00	{}	{}	standard	80dc64ca-4393-4cf3-897b-2ade3321c8a2	2026-07-22 07:37:39.263321	2026-07-22 07:37:31.434656	2026-07-24 07:37:27.658	2026-07-22 07:37:27.805685	2026-07-22 07:37:39.263321	{}	
2a7a15e0-9167-4524-abbd-bfd138bbaa97	CART-20260722-0281	1	1	1	guest-1784701558975	guest@gmail.com	03123456789	converted	standard	pos	\N	\N	{}	4	1	9	20.95	0.00	3.14	0.00	24.09			0.00	{}	{}	standard	51ce494c-ab3a-4854-bd8e-f5a2cb82c3a5	2026-07-22 06:26:08.053801	2026-07-22 06:26:05.42092	2026-07-24 06:25:58.975	2026-07-22 06:25:59.116252	2026-07-22 06:26:08.053801	{}	
baccf90f-7f2c-4974-b935-5754c1f8f28a	CART-20260722-4364	1	1	1	guest-1784705896693	guest@gmail.com	03123456789	converted	standard	pos	\N	\N	{}	4	1	9	425.00	0.00	63.75	0.00	488.75			0.00	{}	{}	standard	e72fdba4-6d6f-4c2a-8337-606bdca5c876	2026-07-22 07:38:22.782073	2026-07-22 07:38:21.303263	2026-07-24 07:38:16.693	2026-07-22 07:38:16.842411	2026-07-22 07:38:22.782073	{}	
6bbea337-0a12-45a9-afb8-0e8eaeeaf9da	CART-20260722-8208	1	1	10	guest-1784714566248	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	8	4.50	0.00	0.67	0.00	5.17			0.00	{}	{}	standard	b5f91a7e-875e-46c5-94d1-c86525cccd5f	2026-07-22 10:02:51.748435	2026-07-22 10:02:48.07015	2026-07-24 10:02:46.248	2026-07-22 10:02:46.382108	2026-07-22 10:02:51.748435	{}	
e52c169d-1614-411f-8bac-ab2db8c3a3af	CART-20260722-7300	1	1	10	guest-1784726097007	Guest@gmail.com	12345678	active	standard	pos	\N	\N	{}	4	1	8	18.95	0.00	2.84	0.00	21.79			0.00	{}	{}	standard	\N	2026-07-22 13:16:21.159	2026-07-22 13:15:06.136563	2026-07-24 13:14:57.007	2026-07-22 13:14:57.143204	2026-07-22 13:16:22.041423	{}	
55ac535e-f21a-4892-b7c4-ec5f824d8976	CART-20260722-4338	1	1	10	guest-1784724616762	Guest@gmail.com	12345678	active	standard	pos	\N	\N	{}	4	1	8	431.40	0.00	32.36	0.00	463.76			0.00	{}	{}	standard	\N	2026-07-22 12:50:41.205	2026-07-22 12:50:27.921625	2026-07-24 12:50:16.762	2026-07-22 12:50:16.857589	2026-07-22 12:50:41.346315	{}	
f3bc32e2-0d76-4bb6-9ed8-7f02b3021566	CART-20260722-2913	1	1	10	guest-1784725826302	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	8	342.20	0.00	51.32	0.00	393.52			0.00	{}	{}	standard	11003230-d7b2-44d1-b057-5a5347af254d	2026-07-22 13:11:05.655009	2026-07-22 13:10:51.259823	2026-07-24 13:10:26.302	2026-07-22 13:10:26.388518	2026-07-22 13:11:05.655009	{}	
6f23c0ff-e3eb-4ba6-9996-2225ed1afb14	CART-20260722-1789	1	1	10	guest-1784726308124	Guest@gmail.com	12345678	active	standard	pos	\N	\N	{}	4	1	8	180.00	0.00	27.00	0.00	207.00			0.00	{}	{}	standard	\N	\N	2026-07-22 13:18:38.19495	2026-07-24 13:18:28.124	2026-07-22 13:18:28.240917	2026-07-22 13:19:11.187794	{}	
1e4579b0-f3ac-4064-ad89-169eb05ffa12	CART-20260727-4008	1	1	10	guest-1785156035548	Guest@gmail.com	12345678	active	standard	pos	\N	\N	{}	4	1	9	4.50	0.00	0.67	0.00	5.17			0.00	{}	{}	standard	\N	\N	2026-07-27 12:40:42.285026	2026-07-29 12:40:35.548	2026-07-27 12:40:35.218289	2026-07-27 12:40:42.288212	{"customer_name": "Guest ضيف"}	
0b90118f-6987-47c4-bc66-3325bd219b63	CART-20260723-5728	1	1	10	guest-1784803946125	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	9	4.50	0.00	0.67	0.00	5.17			0.00	{}	{}	standard	275b4bb0-20a8-430e-be00-7e0ee053f49f	2026-07-23 10:54:04.965803	2026-07-23 10:53:29.421754	2026-07-25 10:52:26.125	2026-07-23 10:52:25.589104	2026-07-23 10:54:04.965803	{}	
deb0a121-45eb-439b-ab8b-b6e4979a960f	CART-20260731-3080	1	1	10	guest-1785493400741	Guest@gmail.com	12345678	active	standard	pos	\N	\N	{}	4	2	8	8.51	0.00	1.27	0.00	9.78			0.00	{}	{}	standard	\N	\N	2026-07-31 10:23:28.811259	2026-08-02 10:23:20.741	2026-07-31 10:23:21.939265	2026-07-31 10:23:28.813923	{"customer_name": "Guest ضيف"}	
fab12c2e-010a-4765-8f17-6e5cdcb041c6	CART-20260723-0182	1	1	10	guest-1784804966140	Guest@gmail.com	12345678	active	standard	pos	\N	\N	{}	4	1	9	12.95	0.00	1.94	0.00	14.89			0.00	{}	{}	standard	\N	2026-07-23 11:09:43.339	2026-07-23 11:09:39.851181	2026-07-25 11:09:26.14	2026-07-23 11:09:25.605093	2026-07-23 11:09:42.808843	{}	
dc47e996-a7df-481b-869d-6f85fb5f51c5	CART-20260803-1627	1	8	\N	guest-1785758022998			active	standard	pos	\N	\N	{}	4	1	12	3.50	0.00	0.53	0.00	4.03			0.00	{}	{}	standard	\N	\N	2026-08-03 11:53:43.345966	2026-08-05 11:53:42.998	2026-08-03 11:53:42.80534	2026-08-03 11:53:43.348731	{"customer_name": "Walk-in Guest"}	
a6a8a674-e57e-4809-9e86-b7d517ef7d4f	CART-20260731-9193	1	8	10	guest-1785483582011	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	12	58.00	0.00	8.73	0.00	66.73			0.00	{}	{}	standard	d7b077c6-d905-4a48-b75d-96acf66acca6	2026-07-31 09:02:44.031443	2026-07-31 09:02:33.183981	2026-08-02 07:39:42.011	2026-07-31 07:39:43.220322	2026-07-31 09:02:44.031443	{"customer_name": "Guest ضيف"}	
7a73337e-1c3b-4c14-9cd9-9e449851d7ee	CART-20260728-2961	1	1	10	guest-1785227823895	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	9	42.45	0.00	6.36	0.00	48.81			0.00	{}	{}	standard	2289e488-034e-4715-803e-93febd00c31b	2026-07-28 08:37:14.651939	2026-07-28 08:37:10.534873	2026-07-30 08:37:03.895	2026-07-28 08:37:02.933244	2026-07-28 08:37:14.651939	{"customer_name": "Guest ضيف"}	
252baaaf-5ae9-4d3a-84fe-3e0878ca2b3f	CART-20260730-8309	1	1	10	guest-1785398973428	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	8	29.45	0.00	4.41	0.00	33.86			0.00	{}	{}	standard	5fd9d2c3-7306-4627-994e-7fbfd52e50a0	2026-07-30 08:10:35.372733	2026-07-30 08:09:36.701344	2026-08-01 08:09:33.428	2026-07-30 08:09:32.423839	2026-07-30 08:10:35.372733	{"customer_name": "Guest ضيف"}	
b627c8a2-bc74-4bb8-9b28-0790aab1266a	CART-20260801-0585	1	1	10	guest-1785565403676	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	2	8	14.00	0.00	2.12	0.00	16.12			0.00	{}	{}	standard	336b87c3-8860-4889-953b-95145ea1aeab	2026-08-01 06:26:48.164436	2026-08-01 06:26:36.490242	2026-08-03 06:23:23.676	2026-08-01 06:23:24.21722	2026-08-01 06:26:48.164436	{"customer_name": "Guest ضيف"}	
3469e4f7-1aef-475b-951b-2cd3ec962029	CART-20260731-1646	1	1	10	guest-1785489204085	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	8	29.45	0.00	4.41	0.00	33.86			0.00	{}	{}	standard	0fcd3efb-e505-45fe-b224-b8010407fd0e	2026-07-31 09:13:36.240484	2026-07-31 09:13:28.828299	2026-08-02 09:13:24.085	2026-07-31 09:13:25.286102	2026-07-31 09:13:36.240484	{"customer_name": "Guest ضيف"}	
e1dad131-4aa2-487a-b854-e002b43febe4	CART-20260727-1520	1	1	10	guest-1785153108176	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	9	269.70	0.00	40.21	0.00	309.91			0.00	{}	{}	standard	4aa69799-b182-41e0-9fba-92c284472094	2026-07-27 11:56:46.137556	2026-07-27 11:54:21.874878	2026-07-29 11:51:48.176	2026-07-27 11:51:47.841451	2026-07-27 11:56:46.137556	{"customer_name": "Guest ضيف"}	
7750e699-09d6-45b6-a904-e408f9dabaef	CART-20260803-7862	1	1	10	guest-1785734518584	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	8	358.20	0.00	53.69	0.00	411.89			0.00	{}	{}	standard	50388992-d30f-4ee7-929d-ee180ede36a7	2026-08-03 05:31:22.813367	2026-08-03 05:30:49.305968	2026-08-05 05:21:58.584	2026-08-03 05:21:57.391925	2026-08-03 05:31:22.813367	{"customer_name": "Guest ضيف"}	
464cd515-42cb-457f-92fe-fb72b21ae341	CART-20260803-7117	1	8	10	guest-1785750968922	Guest@gmail.com	12345678	active	standard	pos	\N	\N	{}	4	1	12	0.00	0.00	0.00	0.00	0.00			0.00	{}	{}	standard	\N	\N	2026-08-03 09:56:08.264297	2026-08-05 09:56:08.922	2026-08-03 09:56:08.264297	2026-08-03 09:56:08.264297	{"customer_name": "Guest ضيف"}	
73419d52-4182-413f-a4da-14cc77155a65	CART-20260803-8550	1	8	10	guest-1785756111396	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	12	18.00	0.00	2.70	0.00	20.70			0.00	{}	{}	standard	480b985a-df97-4c2b-95c3-bef7d77c0e72	2026-08-03 11:39:23.332308	2026-08-03 11:21:55.824097	2026-08-05 11:21:51.396	2026-08-03 11:21:50.722652	2026-08-03 11:39:23.332308	{"customer_name": "Guest ضيف"}	
334043fe-9a22-4e2d-b316-8b780d6717dd	CART-20260727-0138	1	1	10	guest-1785153427925	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	9	724.00	0.00	108.61	0.00	832.61			0.00	{}	{}	standard	f556b2b5-b3c8-4f70-8ccb-cc09f953a861	2026-07-27 12:05:02.734041	2026-07-27 12:02:49.657691	2026-07-29 11:57:07.925	2026-07-27 11:57:07.541632	2026-07-27 12:05:02.734041	{"customer_name": "Guest ضيف"}	
dabc964f-45e9-4234-a23d-4e2fddcdaaee	CART-20260803-2640	1	1	10	guest-1785753356714	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	8	4.50	0.00	0.67	0.00	5.17			0.00	{}	{}	standard	8387e582-d524-4c27-8c35-24ec3f9e8d60	2026-08-03 10:36:05.368178	2026-08-03 10:35:59.745194	2026-08-05 10:35:56.714	2026-08-03 10:35:56.566577	2026-08-03 10:36:05.368178	{"customer_name": "Guest ضيف"}	
6b98ff85-5ee1-493a-9c4f-a1e47cb12075	CART-20260730-4881	1	1	10	guest-1785409408426	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	8	548.42	0.00	82.20	0.00	630.62			0.00	{}	{}	standard	a4e64219-9b7f-45c4-9a35-106788960f26	2026-07-30 11:04:16.88959	2026-07-30 11:04:09.489956	2026-08-01 11:03:28.426	2026-07-30 11:03:27.362417	2026-07-30 11:04:16.88959	{"customer_name": "Guest ضيف"}	
339d0029-bbdf-49c3-8fc1-cadba066f020	CART-20260727-9210	1	1	10	guest-1785153920632	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	9	119.70	0.00	17.95	0.00	137.65			0.00	{}	{}	standard	e983e5b0-8b05-4ace-95b8-b629a31013ac	2026-07-27 12:06:45.369439	2026-07-27 12:06:39.850745	2026-07-29 12:05:20.632	2026-07-27 12:05:20.256804	2026-07-27 12:06:45.369439	{"customer_name": "Guest ضيف"}	
14edb196-5851-4397-b37d-66322afe2a2d	CART-20260727-1224	1	1	10	guest-1785152460201	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	9	56.95	0.00	8.57	0.00	65.52			0.00	{}	{}	standard	358a5da2-fb5c-496d-b582-a1fd790b4535	2026-07-27 11:42:20.346292	2026-07-27 11:41:40.704635	2026-07-29 11:41:00.201	2026-07-27 11:40:59.824434	2026-07-27 11:42:20.346292	{"customer_name": "Guest ضيف"}	
6a6fff8d-1324-4573-9966-fbf179c43933	CART-20260723-8471	1	1	1	guest-1784804987522	guest@gmail.com	03123456789	active	standard	pos	\N	\N	{}	4	1	9	24.95	0.00	3.74	0.00	28.69			0.00	{}	{}	standard	\N	2026-07-23 11:12:12.974	2026-07-23 11:12:10.753556	2026-07-25 11:09:47.522	2026-07-23 11:09:46.971962	2026-07-23 11:12:12.447075	{}	
7335afec-d4d7-4839-bec9-90c3b14ecb53	CART-20260723-9164	1	1	1	guest-1784805143279	guest@gmail.com	03123456789	active	standard	pos	\N	\N	{}	4	1	9	0.00	0.00	0.00	0.00	0.00			0.00	{}	{}	standard	\N	\N	2026-07-23 11:12:22.771801	2026-07-25 11:12:23.279	2026-07-23 11:12:22.771801	2026-07-23 11:12:22.771801	{"customer_name": "aku"}	
2eb60586-650e-4927-a018-945a436f01a3	CART-20260727-3802	1	1	10	guest-1785149238911	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	8	24.95	0.00	3.74	0.00	28.69			0.00	{}	{}	standard	17ca1d02-1e68-4055-8370-93d99ef0a469	2026-07-27 10:47:45.066852	2026-07-27 10:47:21.203472	2026-07-29 10:47:18.911	2026-07-27 10:47:18.54467	2026-07-27 10:47:45.066852	{"customer_name": "Guest ضيف"}	
c269c10a-80e9-41ee-a08b-9c5360c2e4d1	CART-20260723-0405	1	1	10	guest-1784806956703	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	9	4.50	0.00	0.67	0.00	5.17			0.00	{}	{}	standard	75b6b802-225e-4481-8ef4-cabb0bbfd07c	2026-07-23 11:42:46.148176	2026-07-23 11:42:38.83135	2026-07-25 11:42:36.703	2026-07-23 11:42:36.192926	2026-07-23 11:42:46.148176	{"customer_name": "Guest ضيف"}	
3b93a016-e219-4b71-a2a8-9ef8886d2aae	CART-20260731-6070	1	1	10	guest-1785492238591	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	44	6	8	199.11	0.00	29.83	0.00	228.94			0.00	{}	{}	standard	8076b3b4-166f-457d-8358-ba2fd13a3832	2026-07-31 10:19:10.720094	2026-07-31 10:18:42.149971	2026-08-02 10:03:58.591	2026-07-31 10:03:59.817544	2026-07-31 10:19:10.720094	{"customer_name": "Guest ضيف"}	
25f07dc1-0c1b-4fc4-a9a5-da5398c129a7	CART-20260728-1276	1	1	10	guest-1785227744866	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	9	29.90	0.00	4.48	0.00	34.38			0.00	{}	{}	standard	6fd4ff9a-45d4-4ed7-8002-8b58edc2c9ae	2026-07-28 08:35:54.498196	2026-07-28 08:35:51.910496	2026-07-30 08:35:44.866	2026-07-28 08:35:43.919138	2026-07-28 08:35:54.498196	{"customer_name": "Guest ضيف"}	
62161eda-60af-4495-a576-c6d352e0b040	CART-20260730-1106	1	10	10	guest-1785393846746	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	17	113.70	0.00	17.05	0.00	130.75			0.00	{}	{}	standard	27b796e5-3c0e-40f5-810e-027abd4a7ae4	2026-07-30 08:02:53.772947	2026-07-30 08:02:45.658538	2026-08-01 06:44:06.746	2026-07-30 06:44:05.589335	2026-07-30 08:02:53.772947	{"customer_name": "Guest ضيف"}	
c62c4d60-b9e9-4c32-9c09-413b37bed03e	CART-20260727-1056	1	1	10	guest-1785150375823	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	9	29.45	0.00	4.41	0.00	33.86			0.00	{}	{}	standard	2bd4df2f-49f8-409f-aaaa-559039b645fd	2026-07-27 11:06:26.237358	2026-07-27 11:06:18.18815	2026-07-29 11:06:15.823	2026-07-27 11:06:15.443937	2026-07-27 11:06:26.237358	{"customer_name": "Guest ضيف"}	
15c5af82-36c7-46c8-bbdc-9a6c6279d6ae	CART-20260731-1650	1	8	\N	guest-1785496568050			active	standard	pos	\N	\N	{}	4	2	12	0.00	0.00	0.00	0.00	0.00			0.00	{}	{}	standard	\N	\N	2026-07-31 11:16:17.071913	2026-08-02 11:16:08.05	2026-07-31 11:16:09.433274	2026-07-31 11:16:17.073551	{"customer_name": "Walk-in Guest"}	
bfd00d5b-5d48-491b-ade1-b18de6afcb68	CART-20260724-3315	1	1	10	guest-1784884467457	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	9	63.30	0.00	7.47	0.00	70.77			0.00	{}	{}	standard	2b000f45-99b0-4a8a-b4c7-60dcf39af0b6	2026-07-24 09:14:58.90756	2026-07-24 09:14:42.8027	2026-07-26 09:14:27.457	2026-07-24 09:14:28.985359	2026-07-24 09:14:58.90756	{"customer_name": "Guest ضيف"}	
05664a1e-fd9e-4045-81b4-fd5cb0d4110c	CART-20260730-8805	1	1	10	guest-1785409155680	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	8	79.35	0.00	11.89	0.00	91.24			0.00	{}	{}	standard	5a6f336d-d29c-45b1-955e-e19af7441324	2026-07-30 10:59:28.61012	2026-07-30 10:59:22.210962	2026-08-01 10:59:15.68	2026-07-30 10:59:14.603414	2026-07-30 10:59:28.61012	{"customer_name": "Guest ضيف"}	
a2eebeb3-662c-46df-82ef-e6ad81058c5b	CART-20260727-4634	1	1	10	guest-1785150665684	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	9	24.95	0.00	3.74	0.00	28.69			0.00	{}	{}	standard	d11c1e2b-df97-465b-bff1-34b6780ff4fc	2026-07-27 11:11:19.925869	2026-07-27 11:11:14.541821	2026-07-29 11:11:05.684	2026-07-27 11:11:05.394241	2026-07-27 11:11:19.925869	{"customer_name": "Guest ضيف"}	
c37191dd-8749-477b-bbe3-402c2e09a31d	CART-20260801-2980	1	1	10	guest-1785565064699	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	2	8	40.45	0.00	6.07	0.00	46.52			0.00	{}	{}	standard	098814e2-f46e-4be6-9730-87857812e82a	2026-08-01 06:18:24.44564	2026-08-01 06:17:58.694945	2026-08-03 06:17:44.699	2026-08-01 06:17:45.252387	2026-08-01 06:18:24.44564	{"customer_name": "Guest ضيف"}	
de7f9185-027d-4156-bf2c-7b03cd948e88	CART-20260801-3248	1	1	10	guest-1785566791604	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	2	8	28.45	0.00	4.27	0.00	32.72			0.00	{}	{}	standard	a7dcdda8-9253-4201-a387-a0276ab91c39	2026-08-01 06:47:31.587683	2026-08-01 06:46:42.403777	2026-08-03 06:46:31.604	2026-08-01 06:46:32.137222	2026-08-01 06:47:31.587683	{"customer_name": "Guest ضيف"}	
aed43720-0c89-4be2-bec4-cbba48662f56	CART-20260730-6837	1	1	10	guest-1785409688858	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	8	506.38	0.00	75.89	0.00	582.27			0.00	{}	{}	standard	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	2026-07-30 11:08:54.534423	2026-07-30 11:08:51.554629	2026-08-01 11:08:08.858	2026-07-30 11:08:07.767283	2026-07-30 11:08:54.534423	{"customer_name": "Guest ضيف"}	
d397fad3-1a15-4db2-a6db-bec5bf2b9297	CART-20260731-1519	1	8	10	guest-1785488699635		12345678	converted	standard	pos	\N	\N	{}	4	1	12	3.50	0.00	0.53	0.00	4.03			0.00	{}	{}	standard	bcf0f68b-1976-4186-9409-13d858a952f6	2026-07-31 09:05:05.107532	2026-07-31 09:05:01.38503	2026-08-02 09:04:59.635	2026-07-31 09:05:00.848101	2026-07-31 09:05:05.107532	{"customer_name": "Guest ضيف"}	
35aa084d-d355-48f6-88d7-8817ee31408b	CART-20260803-2697	1	1	10	guest-1785753253787	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	8	4.50	0.00	0.67	0.00	5.17			0.00	{}	{}	standard	8fd7a740-50df-4b24-b468-6eabcd6ec92d	2026-08-03 10:34:21.520275	2026-08-03 10:34:16.375017	2026-08-05 10:34:13.787	2026-08-03 10:34:13.128182	2026-08-03 10:34:21.520275	{"customer_name": "Guest ضيف"}	
b014c45e-d107-4546-9852-0ceafe1d6282	CART-20260803-4557	1	8	\N	guest-1785749843379			converted	standard	pos	\N	\N	{}	4	1	12	56.85	0.00	8.52	0.00	65.37			0.00	{}	{}	standard	ffc609d0-0015-4ced-9d82-025d4f8a7ca3	2026-08-03 09:55:00.656239	2026-08-03 09:52:47.351	2026-08-05 09:37:23.379	2026-08-03 09:37:23.297295	2026-08-03 09:55:00.656239	{"customer_name": "Walk-in Guest"}	
8f1bca0d-e6d0-434e-9895-8a68343bd13f	CART-20260731-9962	1	1	10	guest-1785488785058	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	8	63.91	0.00	9.56	0.00	73.47			0.00	{}	{}	standard	e6c4a620-4b91-4708-bec4-40e04779346b	2026-07-31 09:11:02.466443	2026-07-31 09:10:16.705865	2026-08-02 09:06:25.058	2026-07-31 09:06:26.254234	2026-07-31 09:11:02.466443	{"customer_name": "Guest ضيف"}	
9f357046-da05-423c-8753-9c3031e83a44	CART-20260803-3041	1	1	10	guest-1785753380008	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	8	8.00	0.00	1.20	0.00	9.20			0.00	{}	{}	standard	ce52051e-4a6e-440e-909a-86dd0703926d	2026-08-03 10:36:47.658755	2026-08-03 10:36:40.970869	2026-08-05 10:36:20.008	2026-08-03 10:36:19.333169	2026-08-03 10:36:47.658755	{"customer_name": "Guest ضيف"}	
4620d829-c1f1-4885-9686-3bf3796b3cbb	CART-20260803-1389	1	8	10	guest-1785757283760	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	12	3.50	0.00	0.53	0.00	4.03			0.00	{}	{}	standard	332d8ea2-6299-45fc-b833-ad082299dd72	2026-08-03 11:41:39.519443	2026-08-03 11:41:25.108536	2026-08-05 11:41:23.76	2026-08-03 11:41:23.068887	2026-08-03 11:41:39.519443	{"customer_name": "Guest ضيف"}	
a0f2df8d-1907-41dd-927d-749d3e09730b	CART-20260811-4470	1	1	10	guest-1786447426802	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	8	7.00	0.00	1.06	0.00	8.06			0.00	{}	{}	standard	0ad0c821-b76b-4f0c-b149-7edf6df71da9	2026-08-11 11:24:31.117197	2026-08-11 11:24:21.113664	2026-08-13 11:23:46.802	2026-08-11 11:23:46.51765	2026-08-11 11:24:31.117197	{"customer_name": "Guest ضيف"}	
ecaed928-ee99-437d-8e8d-e46d4597aff8	CART-20260811-7936	1	1	10	guest-1786447637428	Guest@gmail.com	12345678	converted	standard	pos	\N	\N	{}	4	1	8	41.50	0.00	6.20	0.00	47.70			0.00	{}	{}	standard	1f34855a-1ff6-4e77-a567-65232d4a4504	2026-08-11 11:46:45.409218	2026-08-11 11:43:14.637721	2026-08-13 11:27:17.428	2026-08-11 11:27:17.131806	2026-08-11 11:46:45.409218	{"customer_name": "Guest ضيف"}	
\.


--
-- Data for Name: cashier_sessions; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.cashier_sessions (id, cashier_id, pos_terminal_id, session_number, opening_time, closing_time, opening_balance, closing_balance, expected_balance, variance, status, metadata, created_at, updated_at) FROM stdin;
29	4	17	SES-20260730-958	2026-07-30 11:57:54.936232	2026-07-30 11:58:04.213622	1000.00	1000.00	1000.00	0.00	closed	{"closed_by": 31, "closing_note": ""}	2026-07-30 11:57:54.936232	2026-07-30 11:58:04.213622
2	1	8	SES-20260720-633	2026-07-20 10:15:56.661946	2026-07-20 10:16:07.266166	1000.00	1000.00	1000.00	0.00	closed	{"closed_by": 4, "closing_note": ""}	2026-07-20 10:15:56.661946	2026-07-20 10:16:07.266166
11	1	8	SES-20260722-741	2026-07-22 13:07:32.646161	2026-07-23 06:14:20.681348	1000.00	2201.04	2201.04	0.00	closed	{"closed_by": 4, "closing_note": ""}	2026-07-22 13:07:32.646161	2026-07-23 06:14:20.681348
3	1	8	SES-20260720-663	2026-07-20 10:19:51.904967	2026-07-20 11:41:16.92924	1000.00	1048.75	1000.00	48.75	closed	{"closed_by": 4, "closing_note": ""}	2026-07-20 10:19:51.904967	2026-07-20 11:41:16.92924
1	1	8	SES-20260720-673	2026-07-20 07:35:31.21447	2026-07-20 11:42:33.040559	1000.00	2281.00	2281.88	-0.88	closed	{"closed_by": 4, "closing_note": ""}	2026-07-20 07:35:31.21447	2026-07-20 11:42:33.040559
21	1	8	SES-20260727-002	2026-07-27 10:46:54.295264	2026-07-27 11:32:55.42039	1000.00	1120.00	1119.93	0.07	closed	{"closed_by": 4, "closing_note": ""}	2026-07-27 10:46:54.295264	2026-07-27 11:32:55.42039
4	1	8	SES-20260720-160	2026-07-20 11:42:58.031075	2026-07-20 11:44:02.101446	1000.00	1065.44	1065.44	0.00	closed	{"closed_by": 4, "closing_note": ""}	2026-07-20 11:42:58.031075	2026-07-20 11:44:02.101446
5	1	12	SES-20260721-680	2026-07-21 06:54:50.012072	2026-07-21 07:05:28.404141	1000.00	1000.00	1000.00	0.00	closed	{"closed_by": 1, "closing_note": "test"}	2026-07-21 06:54:50.012072	2026-07-21 07:05:28.404141
12	1	9	SES-20260723-795	2026-07-23 06:17:34.665149	2026-07-23 06:19:06.789046	1000.00	1094.00	1094.06	-0.06	closed	{"closed_by": 4, "closing_note": ""}	2026-07-23 06:17:34.665149	2026-07-23 06:19:06.789046
30	4	17	SES-20260730-934	2026-07-30 11:59:50.706701	2026-07-30 12:00:06.024377	1000.00	1000.00	1000.00	0.00	closed	{"closed_by": 31, "closing_note": ""}	2026-07-30 11:59:50.706701	2026-07-30 12:00:06.024377
13	1	9	SES-20260723-528	2026-07-23 06:28:10.05599	2026-07-23 06:33:10.541499	1000.00	1070.00	1075.78	-5.78	closed	{"closed_by": 4, "closing_note": ""}	2026-07-23 06:28:10.05599	2026-07-23 06:33:10.541499
7	1	9	SES-20260721-655	2026-07-21 07:11:13.426106	2026-07-21 12:21:30.03123	1999.00	1999.00	1999.00	0.00	closed	{"closed_by": 1, "closing_note": ""}	2026-07-21 07:11:13.426106	2026-07-21 12:21:30.03123
6	1	12	SES-20260721-789	2026-07-21 07:10:13.699176	2026-07-21 12:28:57.227434	1990.00	2860.00	2863.50	-3.50	closed	{"closed_by": 1, "closing_note": ""}	2026-07-21 07:10:13.699176	2026-07-21 12:28:57.227434
14	1	9	SES-20260723-754	2026-07-23 06:33:25.020614	2026-07-23 06:33:45.918814	1000.00	1000.00	1000.00	0.00	closed	{"closed_by": 4, "closing_note": ""}	2026-07-23 06:33:25.020614	2026-07-23 06:33:45.918814
31	1	12	SES-20260731-332	2026-07-31 09:04:58.849048	2026-07-31 09:15:13.990965	1000.00	1140.00	1140.20	-0.20	closed	{"closed_by": 4, "closing_note": ""}	2026-07-31 09:04:58.849048	2026-07-31 09:15:13.990965
15	1	9	SES-20260723-334	2026-07-23 06:37:33.652088	2026-07-23 06:38:40.455726	1000.00	1010.34	1010.34	0.00	closed	{"closed_by": 4, "closing_note": ""}	2026-07-23 06:37:33.652088	2026-07-23 06:38:40.455726
22	1	9	SES-20260727-205	2026-07-27 11:40:42.824916	2026-07-27 12:22:18.160145	1000.00	2345.00	2345.69	-0.69	closed	{"closed_by": 4, "closing_note": ""}	2026-07-27 11:40:42.824916	2026-07-27 12:22:18.160145
8	1	9	SES-20260722-016	2026-07-22 06:25:30.990389	2026-07-22 08:45:53.901592	1000.00	2782.00	2782.14	-0.14	closed	{"closed_by": 4, "closing_note": ""}	2026-07-22 06:25:30.990389	2026-07-22 08:45:53.901592
9	1	8	SES-20260722-333	2026-07-22 08:50:22.259094	2026-07-22 08:52:43.644845	1000.00	1000.00	1000.00	0.00	closed	{"closed_by": 4, "closing_note": ""}	2026-07-22 08:50:22.259094	2026-07-22 08:52:43.644845
16	1	9	SES-20260723-743	2026-07-23 08:30:16.926761	2026-07-23 10:43:15.190415	1000.00	1040.00	1038.98	1.02	closed	{"closed_by": 4, "closing_note": ""}	2026-07-23 08:30:16.926761	2026-07-23 10:43:15.190415
23	1	9	SES-20260727-740	2026-07-27 12:26:37.006279	2026-07-27 12:27:39.623589	1000.00	1000.00	1000.00	0.00	closed	{"closed_by": 4, "closing_note": ""}	2026-07-27 12:26:37.006279	2026-07-27 12:27:39.623589
24	1	9	SES-20260727-367	2026-07-27 12:30:56.589624	2026-07-27 12:31:19.870768	1000.00	1000.00	1000.00	0.00	closed	{"closed_by": 4, "closing_note": ""}	2026-07-27 12:30:56.589624	2026-07-27 12:31:19.870768
25	1	9	SES-20260727-980	2026-07-27 12:40:22.071411	2026-07-28 10:51:18.769778	1000.00	1000.00	1083.19	-83.19	closed	{"closed_by": 1, "closing_note": ""}	2026-07-27 12:40:22.071411	2026-07-28 10:51:18.769778
17	1	9	SES-20260723-780	2026-07-23 10:43:25.37687	2026-07-23 11:25:55.251685	1000.00	2792.20	2792.20	0.00	closed	{"closed_by": 4, "closing_note": ""}	2026-07-23 10:43:25.37687	2026-07-23 11:25:55.251685
18	1	9	SES-20260723-096	2026-07-23 11:42:26.401562	2026-07-23 11:44:15.185198	1000.00	1010.34	1010.34	0.00	closed	{"closed_by": 4, "closing_note": ""}	2026-07-23 11:42:26.401562	2026-07-23 11:44:15.185198
10	1	8	SES-20260722-224	2026-07-22 08:53:45.327134	2026-07-22 13:03:45.702439	1000.00	3530.00	3529.82	0.18	closed	{"closed_by": 4, "closing_note": ""}	2026-07-22 08:53:45.327134	2026-07-22 13:03:45.702439
19	1	8	SES-20260723-066	2026-07-23 11:46:04.839442	2026-07-23 11:46:15.001681	1000.00	1000.00	1000.00	0.00	closed	{"closed_by": 4, "closing_note": ""}	2026-07-23 11:46:04.839442	2026-07-23 11:46:15.001681
26	1	17	SES-20260730-949	2026-07-30 06:16:30.839265	2026-07-30 08:03:20.561692	1000.00	1130.00	1130.75	-0.75	closed	{"closed_by": 4, "closing_note": ""}	2026-07-30 06:16:30.839265	2026-07-30 08:03:20.561692
20	1	9	SES-20260724-076	2026-07-24 09:14:20.952292	2026-07-27 10:45:58.056468	10000.00	10141.00	10141.54	-0.54	closed	{"closed_by": 4, "closing_note": ""}	2026-07-24 09:14:20.952292	2026-07-27 10:45:58.056468
28	1	8	SES-20260730-746	2026-07-30 11:03:17.300454	2026-07-31 06:56:25.428001	1000.00	2212.00	2212.89	-0.89	closed	{"closed_by": 4, "closing_note": ""}	2026-07-30 11:03:17.300454	2026-07-31 06:56:25.428001
33	1	18	SES-20260731-511	2026-07-31 09:14:29.456434	2026-07-31 07:06:39.987743	1230.00	1239.00	1230.00	9.00	closed	{"closed_by": 4, "closing_note": ""}	2026-07-31 09:14:29.456434	2026-07-31 07:06:39.987743
27	1	8	SES-20260730-846	2026-07-30 08:09:16.757289	2026-07-30 11:01:53.644115	1000.00	1125.00	1125.10	-0.10	closed	{"closed_by": 4, "closing_note": ""}	2026-07-30 08:09:16.757289	2026-07-30 11:01:53.644115
34	6	8	SES-20260731-414	2026-07-31 10:03:45.628277	\N	1000.00	\N	1228.94	\N	open	{}	2026-07-31 10:03:45.628277	2026-07-31 10:19:10.760242
35	2	8	SES-20260731-660	2026-07-31 10:21:33.379644	\N	1000.00	\N	1095.36	\N	open	{}	2026-07-31 10:21:33.379644	2026-08-01 06:47:31.605431
32	1	12	SES-20260731-360	2026-07-31 09:09:48.657929	2026-07-31 09:14:05.800763	1000.00	1038.00	1037.89	0.11	closed	{"closed_by": 4, "closing_note": ""}	2026-07-31 09:09:48.657929	2026-07-31 09:14:05.800763
36	1	8	SES-20260731-016	2026-07-31 15:09:14.41218	2026-07-31 12:15:22.137844	1000.00	1000.00	1000.00	0.00	closed	{"closed_by": 1, "closing_note": ""}	2026-07-31 15:09:14.41218	2026-07-31 12:15:22.137844
37	1	8	SES-20260803-192	2026-08-03 05:21:39.390569	2026-08-03 10:38:28.654716	1000.00	1516.00	1516.34	-0.34	closed	{"closed_by": 4, "closing_note": ""}	2026-08-03 05:21:39.390569	2026-08-03 10:38:28.654716
38	1	12	SES-20260803-758	2026-08-03 11:18:46.69827	2026-08-11 08:12:27.930673	1000.00	1000.00	1049.46	-49.46	closed	{"closed_by": 4, "closing_note": ""}	2026-08-03 11:18:46.69827	2026-08-11 08:12:27.930673
39	1	8	SES-20260811-304	2026-08-11 10:30:33.569992	\N	1000.00	\N	1111.52	\N	open	{}	2026-08-11 10:30:33.569992	2026-08-11 11:46:45.446992
\.


--
-- Data for Name: cashiers; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.cashiers (id, user_id, store_id, cashier_code, drawer_limit, discount_limit, is_active, metadata, created_at) FROM stdin;
1	4	1	CASH-001	5000.00	10.00	t	{}	2026-07-18 07:59:38.038245
2	4	9	CASH-EMP004	1000.00	50.00	t	{}	2026-07-21 09:51:35.020775
3	12	8	CASH01	1000.00	0.00	t	{}	2026-07-24 07:22:41.028735
4	31	1	cash-004	5000.00	10.00	t	{"assigned_terminal_ids": [8, 9]}	2026-07-28 05:45:52.629934
5	44	10	cash1	0.00	0.00	t	{"assigned_terminal_ids": []}	2026-07-30 10:40:11.635699
6	44	17	cash1	1000.00	50.00	t	{}	2026-07-30 11:42:55.904941
\.


--
-- Data for Name: combo_bundle_items; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.combo_bundle_items (id, combo_bundle_id, menu_item_id, product_id, product_variant_id, item_type, quantity, is_required, group_tag, price_override, display_order, metadata, created_at) FROM stdin;
\.


--
-- Data for Name: combo_bundles; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.combo_bundles (id, store_id, code, name, description, bundle_price, bundle_type, is_active, valid_from, valid_to, display_order, metadata, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: customers; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.customers (id, organization_id, customer_code, name, email, phone, address, customer_type, price_list_id, credit_limit, outstanding_balance, loyalty_points, is_active, metadata, created_at, updated_at) FROM stdin;
1	1	GUEST-1784532999686	aku	guest@gmail.com	03123456789	BarberaCafeTabuk,Saudi Arabia	retail	1	1000.00	0.00	0.00	t	{}	2026-07-20 07:36:39.283091	2026-07-22 11:13:14.016297
10	1	hello122	Guest ضيف	Guest@gmail.com	12345678	Guest 123 St	guest	1	1000.00	0.00	0.00	t	{}	2026-07-22 09:13:19.561923	2026-07-22 11:39:53.787135
2	1	CUST001	Jane Doe	jane@example.com	\N	123 Main St	promotional	3	10000.00	0.00	0.00	t	{}	2026-07-20 13:20:12.181992	2026-07-24 17:45:53.248215
\.


--
-- Data for Name: discount_analytics; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.discount_analytics (id, organization_id, store_id, cashier_id, product_id, discount_type, date, month, quarter, year, total_discounts_given, transactions_with_discount, total_transactions, discount_percentage, revenue_impact, metadata, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: draft_cart_template_items; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.draft_cart_template_items (id, template_id, organization_id, product_id, product_variant_id, quantity, uom_id, last_known_price, priority, notes, metadata, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: draft_cart_templates; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.draft_cart_templates (id, organization_id, customer_id, template_name, description, template_type, is_favorite, auto_reorder_enabled, reorder_frequency_days, next_reorder_date, total_items, estimated_total, metadata, notes, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: inventory_analytics; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.inventory_analytics (id, organization_id, store_id, product_id, category_id, date, month, quarter, year, opening_stock, stock_in, stock_out, receipts, issues, adjustments, closing_stock, average_stock, stock_value, turnover_rate, stock_turnover_ratio, days_of_inventory, days_in_stock, low_stock_alerts, out_of_stock_days, metadata, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: inventory_stock; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.inventory_stock (id, product_id, product_variant_id, store_id, storage_location_id, quantity_on_hand, quantity_allocated, quantity_available, quantity_on_order, quantity_in_transit, reorder_level, reorder_quantity, max_stock_level, last_counted_at, metadata, created_at, updated_at) FROM stdin;
210	76	\N	4	9	500.000	50.000	450.000	200.000	100.000	100.000	300.000	1000.000	\N	{}	2026-07-30 11:34:46.014383	2026-07-30 11:34:46.014383
4	4	\N	1	1	100.000	0.000	100.000	0.000	0.000	25.000	80.000	180.000	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
34	34	\N	1	5	-2.000	1.000	0.000	0.000	0.000	20.000	65.000	130.000	\N	{}	2026-07-18 07:59:03.504911	2026-08-03 05:31:30.355697
27	27	\N	1	3	77.000	0.000	77.000	0.000	0.000	22.000	70.000	140.000	\N	{}	2026-07-18 07:59:03.504911	2026-08-03 05:31:30.355697
12	12	\N	1	2	480.000	0.000	480.000	0.000	0.000	120.000	400.000	800.000	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
18	18	\N	1	2	70.000	0.000	70.000	0.000	0.000	18.000	60.000	120.000	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
10	10	\N	1	2	239.000	0.000	239.000	0.000	0.000	60.000	200.000	450.000	\N	{}	2026-07-18 07:59:03.504911	2026-07-30 11:08:58.292688
17	17	\N	1	2	50.000	0.000	50.000	0.000	0.000	12.000	45.000	90.000	\N	{}	2026-07-18 07:59:03.504911	2026-07-30 11:08:58.292688
28	28	\N	1	3	100.000	0.000	100.000	0.000	0.000	25.000	80.000	180.000	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
19	19	\N	1	2	64.000	0.000	64.000	0.000	0.000	16.000	55.000	110.000	\N	{}	2026-07-18 07:59:03.504911	2026-07-30 11:08:58.292688
43	2	\N	2	\N	48.000	0.000	48.000	0.000	0.000	12.000	48.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
44	3	\N	2	\N	36.000	0.000	36.000	0.000	0.000	9.000	36.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
45	4	\N	2	\N	60.000	0.000	60.000	0.000	0.000	15.000	48.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
46	5	\N	2	\N	90.000	0.000	90.000	0.000	0.000	24.000	72.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
47	6	\N	2	\N	54.000	0.000	54.000	0.000	0.000	12.000	42.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
48	7	\N	2	\N	45.000	0.000	45.000	0.000	0.000	9.000	36.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
49	8	\N	2	\N	30.000	0.000	30.000	0.000	0.000	6.000	24.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
50	9	\N	2	\N	173.000	0.000	173.000	0.000	0.000	43.000	144.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
51	10	\N	2	\N	144.000	0.000	144.000	0.000	0.000	36.000	120.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
52	11	\N	2	\N	43.000	0.000	43.000	0.000	0.000	12.000	36.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
53	12	\N	2	\N	288.000	0.000	288.000	0.000	0.000	72.000	240.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
54	13	\N	2	\N	216.000	0.000	216.000	0.000	0.000	54.000	180.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
55	14	\N	2	\N	60.000	0.000	60.000	0.000	0.000	15.000	48.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
56	15	\N	2	\N	51.000	0.000	51.000	0.000	0.000	12.000	42.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
57	16	\N	2	\N	36.000	0.000	36.000	0.000	0.000	9.000	30.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
58	17	\N	2	\N	33.000	0.000	33.000	0.000	0.000	7.000	27.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
59	18	\N	2	\N	42.000	0.000	42.000	0.000	0.000	11.000	36.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
60	19	\N	2	\N	39.000	0.000	39.000	0.000	0.000	10.000	33.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
61	20	\N	2	\N	48.000	0.000	48.000	0.000	0.000	12.000	36.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
62	21	\N	2	\N	27.000	0.000	27.000	0.000	0.000	6.000	24.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
63	22	\N	2	\N	54.000	0.000	54.000	0.000	0.000	15.000	42.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
64	23	\N	2	\N	45.000	0.000	45.000	0.000	0.000	12.000	36.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
65	24	\N	2	\N	72.000	0.000	72.000	0.000	0.000	18.000	60.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
66	25	\N	2	\N	66.000	0.000	66.000	0.000	0.000	17.000	54.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
67	26	\N	2	\N	57.000	0.000	57.000	0.000	0.000	14.000	48.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
68	27	\N	2	\N	51.000	0.000	51.000	0.000	0.000	13.000	42.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
69	28	\N	2	\N	60.000	0.000	60.000	0.000	0.000	15.000	48.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
70	29	\N	2	\N	48.000	0.000	48.000	0.000	0.000	12.000	39.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
71	30	\N	2	\N	42.000	0.000	42.000	0.000	0.000	11.000	36.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
72	31	\N	2	\N	39.000	0.000	39.000	0.000	0.000	10.000	33.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
73	32	\N	2	\N	30.000	0.000	30.000	0.000	0.000	7.000	27.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
74	33	\N	2	\N	90.000	0.000	90.000	0.000	0.000	24.000	72.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
75	34	\N	2	\N	48.000	0.000	48.000	0.000	0.000	12.000	39.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
76	35	\N	2	\N	84.000	0.000	84.000	0.000	0.000	21.000	66.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
77	36	\N	2	\N	60.000	0.000	60.000	0.000	0.000	15.000	51.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
78	37	\N	2	\N	36.000	0.000	36.000	0.000	0.000	9.000	30.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
154	31	\N	4	\N	120.000	0.000	120.000	0.000	0.000	32.000	110.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-30 07:30:33.454023
42	1	\N	2	\N	77.000	0.000	77.000	0.000	0.000	18.000	60.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-28 13:17:43.081608
153	30	\N	4	\N	120.000	0.000	120.000	0.000	0.000	36.000	120.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-30 07:30:33.454023
16	16	\N	1	2	59.000	0.000	59.000	0.000	0.000	15.000	50.000	100.000	\N	{}	2026-07-18 07:59:03.504911	2026-07-30 11:08:58.292688
13	13	\N	1	2	359.000	0.000	359.000	0.000	0.000	90.000	300.000	600.000	\N	{}	2026-07-18 07:59:03.504911	2026-07-30 11:08:58.292688
29	29	\N	1	3	67.000	0.000	67.000	0.000	0.000	20.000	65.000	140.000	\N	{}	2026-07-18 07:59:03.504911	2026-07-30 11:08:58.292688
79	38	\N	2	\N	33.000	0.000	33.000	0.000	0.000	8.000	27.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
80	39	\N	2	\N	30.000	0.000	30.000	0.000	0.000	7.000	24.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
81	40	\N	2	\N	27.000	0.000	27.000	0.000	0.000	7.000	23.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
82	41	\N	2	\N	51.000	0.000	51.000	0.000	0.000	13.000	42.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
83	1	\N	3	\N	48.000	0.000	48.000	0.000	0.000	12.000	40.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
84	2	\N	3	\N	32.000	0.000	32.000	0.000	0.000	8.000	32.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
85	3	\N	3	\N	24.000	0.000	24.000	0.000	0.000	6.000	24.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
86	4	\N	3	\N	40.000	0.000	40.000	0.000	0.000	10.000	32.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
87	5	\N	3	\N	60.000	0.000	60.000	0.000	0.000	16.000	48.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
88	6	\N	3	\N	36.000	0.000	36.000	0.000	0.000	8.000	28.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
89	7	\N	3	\N	30.000	0.000	30.000	0.000	0.000	6.000	24.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
90	8	\N	3	\N	20.000	0.000	20.000	0.000	0.000	4.000	16.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
91	9	\N	3	\N	115.000	0.000	115.000	0.000	0.000	29.000	96.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
92	10	\N	3	\N	96.000	0.000	96.000	0.000	0.000	24.000	80.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
93	11	\N	3	\N	29.000	0.000	29.000	0.000	0.000	8.000	24.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
94	12	\N	3	\N	192.000	0.000	192.000	0.000	0.000	48.000	160.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
95	13	\N	3	\N	144.000	0.000	144.000	0.000	0.000	36.000	120.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
96	14	\N	3	\N	40.000	0.000	40.000	0.000	0.000	10.000	32.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
97	15	\N	3	\N	34.000	0.000	34.000	0.000	0.000	8.000	28.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
98	16	\N	3	\N	24.000	0.000	24.000	0.000	0.000	6.000	20.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
99	17	\N	3	\N	22.000	0.000	22.000	0.000	0.000	5.000	18.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
100	18	\N	3	\N	28.000	0.000	28.000	0.000	0.000	7.000	24.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
101	19	\N	3	\N	26.000	0.000	26.000	0.000	0.000	6.000	22.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
102	20	\N	3	\N	32.000	0.000	32.000	0.000	0.000	8.000	24.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
103	21	\N	3	\N	18.000	0.000	18.000	0.000	0.000	4.000	16.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
104	22	\N	3	\N	36.000	0.000	36.000	0.000	0.000	10.000	28.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
26	26	\N	1	3	75.000	2.000	73.000	0.000	0.000	24.000	80.000	160.000	\N	{}	2026-07-18 07:59:03.504911	2026-08-11 11:46:54.854467
105	23	\N	3	\N	30.000	0.000	30.000	0.000	0.000	8.000	24.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
106	24	\N	3	\N	48.000	0.000	48.000	0.000	0.000	12.000	40.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
107	25	\N	3	\N	44.000	0.000	44.000	0.000	0.000	11.000	36.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
108	26	\N	3	\N	38.000	0.000	38.000	0.000	0.000	10.000	32.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
109	27	\N	3	\N	34.000	0.000	34.000	0.000	0.000	9.000	28.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
110	28	\N	3	\N	40.000	0.000	40.000	0.000	0.000	10.000	32.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
111	29	\N	3	\N	32.000	0.000	32.000	0.000	0.000	8.000	26.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
112	30	\N	3	\N	28.000	0.000	28.000	0.000	0.000	7.000	24.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
113	31	\N	3	\N	26.000	0.000	26.000	0.000	0.000	6.000	22.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
114	32	\N	3	\N	20.000	0.000	20.000	0.000	0.000	5.000	18.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
115	33	\N	3	\N	60.000	0.000	60.000	0.000	0.000	16.000	48.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
116	34	\N	3	\N	32.000	0.000	32.000	0.000	0.000	8.000	26.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
117	35	\N	3	\N	56.000	0.000	56.000	0.000	0.000	14.000	44.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
118	36	\N	3	\N	40.000	0.000	40.000	0.000	0.000	10.000	34.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
119	37	\N	3	\N	24.000	0.000	24.000	0.000	0.000	6.000	20.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
120	38	\N	3	\N	22.000	0.000	22.000	0.000	0.000	6.000	18.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
121	39	\N	3	\N	20.000	0.000	20.000	0.000	0.000	5.000	16.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
122	40	\N	3	\N	18.000	0.000	18.000	0.000	0.000	4.000	15.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
123	41	\N	3	\N	34.000	0.000	34.000	0.000	0.000	9.000	28.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
124	1	\N	4	\N	240.000	0.000	240.000	0.000	0.000	60.000	200.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
125	2	\N	4	\N	160.000	0.000	160.000	0.000	0.000	40.000	160.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
126	3	\N	4	\N	120.000	0.000	120.000	0.000	0.000	30.000	120.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
127	4	\N	4	\N	200.000	0.000	200.000	0.000	0.000	50.000	160.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
128	5	\N	4	\N	300.000	0.000	300.000	0.000	0.000	80.000	240.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
129	6	\N	4	\N	180.000	0.000	180.000	0.000	0.000	40.000	140.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
130	7	\N	4	\N	150.000	0.000	150.000	0.000	0.000	30.000	120.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
131	8	\N	4	\N	100.000	0.000	100.000	0.000	0.000	20.000	80.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
133	10	\N	4	\N	480.000	0.000	480.000	0.000	0.000	120.000	400.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
135	12	\N	4	\N	960.000	0.000	960.000	0.000	0.000	240.000	800.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
136	13	\N	4	\N	720.000	0.000	720.000	0.000	0.000	180.000	600.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
137	14	\N	4	\N	200.000	0.000	200.000	0.000	0.000	50.000	160.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
138	15	\N	4	\N	170.000	0.000	170.000	0.000	0.000	40.000	140.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
139	16	\N	4	\N	120.000	0.000	120.000	0.000	0.000	30.000	100.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
140	17	\N	4	\N	110.000	0.000	110.000	0.000	0.000	24.000	90.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
141	18	\N	4	\N	140.000	0.000	140.000	0.000	0.000	36.000	120.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
142	19	\N	4	\N	130.000	0.000	130.000	0.000	0.000	32.000	110.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
143	20	\N	4	\N	160.000	0.000	160.000	0.000	0.000	40.000	120.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
144	21	\N	4	\N	90.000	0.000	90.000	0.000	0.000	20.000	80.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
145	22	\N	4	\N	180.000	0.000	180.000	0.000	0.000	50.000	140.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
211	77	\N	10	13	500.000	50.000	450.000	200.000	100.000	100.000	200.000	1000.000	\N	{}	2026-07-30 12:05:29.252318	2026-07-30 12:05:29.252318
147	24	\N	4	\N	240.000	0.000	240.000	0.000	0.000	60.000	200.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
148	25	\N	4	\N	220.000	0.000	220.000	0.000	0.000	56.000	180.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
149	26	\N	4	\N	190.000	0.000	190.000	0.000	0.000	48.000	160.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
150	27	\N	4	\N	170.000	0.000	170.000	0.000	0.000	44.000	140.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
151	28	\N	4	\N	200.000	0.000	200.000	0.000	0.000	50.000	160.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
152	29	\N	4	\N	160.000	0.000	160.000	0.000	0.000	40.000	130.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
155	32	\N	4	\N	100.000	0.000	100.000	0.000	0.000	24.000	90.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
156	33	\N	4	\N	300.000	0.000	300.000	0.000	0.000	80.000	240.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
157	34	\N	4	\N	160.000	0.000	160.000	0.000	0.000	40.000	130.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
158	35	\N	4	\N	280.000	0.000	280.000	0.000	0.000	70.000	220.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
159	36	\N	4	\N	200.000	0.000	200.000	0.000	0.000	50.000	170.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
146	23	\N	4	\N	139.000	0.000	139.000	0.000	0.000	40.000	120.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-30 07:47:50.151711
160	37	\N	4	\N	120.000	0.000	120.000	0.000	0.000	30.000	100.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
161	38	\N	4	\N	110.000	0.000	110.000	0.000	0.000	28.000	90.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
162	39	\N	4	\N	100.000	0.000	100.000	0.000	0.000	24.000	80.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
163	40	\N	4	\N	90.000	0.000	90.000	0.000	0.000	22.000	76.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
164	41	\N	4	\N	170.000	0.000	170.000	0.000	0.000	44.000	140.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
165	1	\N	5	\N	180.000	0.000	180.000	0.000	0.000	45.000	150.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
166	2	\N	5	\N	120.000	0.000	120.000	0.000	0.000	30.000	120.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
167	3	\N	5	\N	90.000	0.000	90.000	0.000	0.000	23.000	90.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
168	4	\N	5	\N	150.000	0.000	150.000	0.000	0.000	38.000	120.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
169	5	\N	5	\N	225.000	0.000	225.000	0.000	0.000	60.000	180.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
170	6	\N	5	\N	135.000	0.000	135.000	0.000	0.000	30.000	105.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
171	7	\N	5	\N	113.000	0.000	113.000	0.000	0.000	23.000	90.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
172	8	\N	5	\N	75.000	0.000	75.000	0.000	0.000	15.000	60.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
173	9	\N	5	\N	432.000	0.000	432.000	0.000	0.000	108.000	360.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
174	10	\N	5	\N	360.000	0.000	360.000	0.000	0.000	90.000	300.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
175	11	\N	5	\N	108.000	0.000	108.000	0.000	0.000	30.000	90.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
176	12	\N	5	\N	720.000	0.000	720.000	0.000	0.000	180.000	600.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
177	13	\N	5	\N	540.000	0.000	540.000	0.000	0.000	135.000	450.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
178	14	\N	5	\N	150.000	0.000	150.000	0.000	0.000	38.000	120.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
179	15	\N	5	\N	128.000	0.000	128.000	0.000	0.000	30.000	105.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
180	16	\N	5	\N	90.000	0.000	90.000	0.000	0.000	23.000	75.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
181	17	\N	5	\N	83.000	0.000	83.000	0.000	0.000	18.000	68.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
182	18	\N	5	\N	105.000	0.000	105.000	0.000	0.000	27.000	90.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
183	19	\N	5	\N	98.000	0.000	98.000	0.000	0.000	24.000	83.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
184	20	\N	5	\N	120.000	0.000	120.000	0.000	0.000	30.000	90.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
185	21	\N	5	\N	68.000	0.000	68.000	0.000	0.000	15.000	60.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
186	22	\N	5	\N	135.000	0.000	135.000	0.000	0.000	38.000	105.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
132	9	\N	4	\N	380.000	0.000	380.000	0.000	0.000	144.000	480.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-08-11 09:16:51.275576
187	23	\N	5	\N	113.000	0.000	113.000	0.000	0.000	30.000	90.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
188	24	\N	5	\N	180.000	0.000	180.000	0.000	0.000	45.000	150.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
189	25	\N	5	\N	165.000	0.000	165.000	0.000	0.000	42.000	135.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
190	26	\N	5	\N	143.000	0.000	143.000	0.000	0.000	36.000	120.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
191	27	\N	5	\N	128.000	0.000	128.000	0.000	0.000	33.000	105.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
192	28	\N	5	\N	150.000	0.000	150.000	0.000	0.000	38.000	120.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
193	29	\N	5	\N	120.000	0.000	120.000	0.000	0.000	30.000	98.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
194	30	\N	5	\N	105.000	0.000	105.000	0.000	0.000	27.000	90.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
195	31	\N	5	\N	98.000	0.000	98.000	0.000	0.000	24.000	83.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
196	32	\N	5	\N	75.000	0.000	75.000	0.000	0.000	18.000	68.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
197	33	\N	5	\N	225.000	0.000	225.000	0.000	0.000	60.000	180.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
198	34	\N	5	\N	120.000	0.000	120.000	0.000	0.000	30.000	98.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
199	35	\N	5	\N	210.000	0.000	210.000	0.000	0.000	53.000	165.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
200	36	\N	5	\N	150.000	0.000	150.000	0.000	0.000	38.000	128.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
201	37	\N	5	\N	90.000	0.000	90.000	0.000	0.000	23.000	75.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
202	38	\N	5	\N	83.000	0.000	83.000	0.000	0.000	21.000	68.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
203	39	\N	5	\N	75.000	0.000	75.000	0.000	0.000	18.000	60.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
204	40	\N	5	\N	68.000	0.000	68.000	0.000	0.000	17.000	57.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
205	41	\N	5	\N	128.000	0.000	128.000	0.000	0.000	33.000	105.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
212	34	\N	17	12	50.000	0.000	50.000	0.000	0.000	\N	\N	\N	\N	{}	2026-07-31 06:17:59.038853	2026-07-31 06:18:20.713763
6	6	\N	1	1	79.000	1.000	78.000	0.000	0.000	20.000	70.000	150.000	\N	{}	2026-07-18 07:59:03.504911	2026-07-31 10:19:16.186314
37	37	\N	1	6	57.000	0.000	57.000	0.000	0.000	15.000	50.000	100.000	\N	{}	2026-07-18 07:59:03.504911	2026-07-31 10:19:16.186314
23	23	\N	1	3	65.000	0.000	65.000	0.000	11.000	20.000	60.000	130.000	\N	{}	2026-07-18 07:59:03.504911	2026-08-03 05:31:30.355697
21	21	\N	1	3	39.000	0.000	39.000	0.000	0.000	10.000	40.000	80.000	\N	{}	2026-07-18 07:59:03.504911	2026-07-30 11:04:24.984511
1	1	\N	1	1	113.000	0.000	113.000	0.000	0.000	30.000	100.000	200.000	\N	{}	2026-07-18 07:59:03.504911	2026-07-30 11:08:58.292688
32	32	\N	1	4	48.000	0.000	48.000	0.000	0.000	12.000	45.000	90.000	\N	{}	2026-07-18 07:59:03.504911	2026-07-30 11:08:58.292688
14	14	\N	1	2	32.000	0.000	32.000	0.000	0.000	25.000	80.000	180.000	\N	{}	2026-07-18 07:59:03.504911	2026-07-30 11:08:58.292688
15	15	\N	1	2	82.000	0.000	82.000	0.000	0.000	20.000	70.000	150.000	\N	{}	2026-07-18 07:59:03.504911	2026-07-30 11:08:58.292688
39	39	\N	1	6	48.000	0.000	48.000	0.000	0.000	12.000	40.000	85.000	\N	{}	2026-07-18 07:59:03.504911	2026-07-30 11:08:58.292688
38	38	\N	1	6	53.000	0.000	53.000	0.000	0.000	14.000	45.000	90.000	\N	{}	2026-07-18 07:59:03.504911	2026-07-30 11:08:58.292688
36	36	\N	1	5	98.000	0.000	98.000	0.000	0.000	25.000	85.000	170.000	\N	{}	2026-07-18 07:59:03.504911	2026-07-30 11:08:58.292688
25	25	\N	1	3	108.000	0.000	108.000	0.000	0.000	28.000	90.000	180.000	\N	{}	2026-07-18 07:59:03.504911	2026-07-30 11:08:58.292688
24	24	\N	1	3	115.000	0.000	115.000	0.000	0.000	30.000	100.000	200.000	\N	{}	2026-07-18 07:59:03.504911	2026-07-30 11:08:58.292688
20	20	\N	1	3	78.000	0.000	78.000	0.000	0.000	20.000	60.000	150.000	\N	{}	2026-07-18 07:59:03.504911	2026-07-30 11:08:58.292688
11	11	\N	1	2	70.000	0.000	70.000	0.000	0.000	20.000	60.000	120.000	\N	{}	2026-07-18 07:59:03.504911	2026-07-30 11:08:58.292688
3	3	\N	1	1	57.000	6.000	51.000	0.000	0.000	15.000	60.000	120.000	\N	{}	2026-07-18 07:59:03.504911	2026-07-31 10:19:16.186314
31	31	\N	1	4	72.000	0.000	72.000	0.000	0.000	16.000	55.000	110.000	\N	{}	2026-07-18 07:59:03.504911	2026-07-31 10:19:16.186314
30	30	\N	1	4	86.000	0.000	86.000	0.000	0.000	18.000	60.000	120.000	\N	{}	2026-07-18 07:59:03.504911	2026-07-31 10:19:16.186314
41	41	\N	1	6	82.000	0.000	82.000	0.000	0.000	22.000	70.000	140.000	\N	{}	2026-07-18 07:59:03.504911	2026-07-31 10:19:16.186314
2	2	\N	1	1	76.000	1.000	75.000	0.000	0.000	20.000	80.000	150.000	\N	{}	2026-07-18 07:59:03.504911	2026-07-31 10:19:16.186314
22	22	\N	1	3	54.000	39.000	15.000	0.000	0.000	25.000	70.000	150.000	\N	{}	2026-07-18 07:59:03.504911	2026-08-01 06:18:28.175551
8	8	\N	1	1	37.000	20.000	17.000	0.000	0.000	10.000	40.000	100.000	\N	{}	2026-07-18 07:59:03.504911	2026-08-01 06:18:28.175551
40	40	\N	1	6	12.000	0.000	12.000	0.000	0.000	11.000	38.000	75.000	\N	{}	2026-07-18 07:59:03.504911	2026-08-11 07:20:20.748627
213	7	\N	17	12	20.000	0.000	20.000	0.000	0.000	\N	\N	\N	\N	{}	2026-08-11 07:20:20.748627	2026-08-11 07:20:32.513374
214	40	\N	17	12	2.000	0.000	2.000	0.000	0.000	\N	\N	\N	\N	{}	2026-08-11 07:20:20.748627	2026-08-11 07:20:32.513374
5	5	\N	1	1	140.000	0.000	140.000	0.000	0.000	40.000	120.000	250.000	\N	{}	2026-07-18 07:59:03.504911	2026-08-11 07:26:24.241143
215	5	\N	17	12	9.000	0.000	9.000	0.000	0.000	\N	\N	\N	\N	{}	2026-08-11 07:26:24.241143	2026-08-11 07:26:28.923699
134	11	\N	4	\N	142.000	0.000	142.000	0.000	0.000	40.000	120.000	\N	\N	{}	2026-07-18 07:59:03.504911	2026-08-11 07:28:58.022557
216	11	\N	17	12	2.000	0.000	2.000	0.000	0.000	\N	\N	\N	\N	{}	2026-08-11 07:28:58.022557	2026-08-11 07:29:04.819316
9	9	\N	1	2	363.000	0.000	363.000	0.000	48.000	72.000	240.000	500.000	\N	{}	2026-07-18 07:59:03.504911	2026-08-11 09:17:02.166042
221	9	\N	17	12	72.000	0.000	72.000	0.000	0.000	\N	\N	\N	\N	{}	2026-08-11 08:11:26.183267	2026-08-11 08:11:33.406704
35	35	\N	1	5	61.000	0.000	61.000	0.000	0.000	35.000	110.000	230.000	\N	{}	2026-07-18 07:59:03.504911	2026-08-11 11:46:54.854467
33	33	\N	1	5	72.000	4.000	68.000	0.000	0.000	40.000	120.000	250.000	\N	{}	2026-07-18 07:59:03.504911	2026-08-11 11:46:54.854467
7	7	\N	1	1	50.000	0.000	50.000	0.000	0.000	15.000	60.000	120.000	\N	{}	2026-07-18 07:59:03.504911	2026-08-11 11:46:54.854467
224	79	1	1	1	800.000	100.000	400.000	100.000	200.000	100.000	400.000	1000.000	\N	{}	2026-08-11 12:17:46.14824	2026-08-11 12:17:46.14824
\.


--
-- Data for Name: invoice_lines; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.invoice_lines (id, invoice_id, organization_id, line_number, description, item_type, product_id, product_variant_id, product_sku, order_line_id, quantity, unit_price, discount_amount, tax_amount, line_total, tax_category_id, tax_rate, uom_id, metadata, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: invoice_payments; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.invoice_payments (id, invoice_id, organization_id, payment_number, payment_date, payment_amount, payment_method, payment_gateway, payment_reference, currency_code, exchange_rate, bank_account_id, reconciled, reconciled_date, notes, received_by_user_id, metadata, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: invoice_status_history; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.invoice_status_history (id, invoice_id, organization_id, from_status, to_status, reason, notes, changed_by_user_id, changed_at) FROM stdin;
\.


--
-- Data for Name: invoices; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.invoices (id, invoice_number, organization_id, store_id, customer_id, customer_name, customer_email, customer_phone, customer_tax_id, invoice_type, invoice_status, sales_order_id, related_invoice_id, invoice_date, due_date, sent_date, paid_date, subtotal, discount_amount, tax_amount, shipping_amount, adjustment_amount, total_amount, paid_amount, credit_applied, balance_due, payment_terms, currency_code, exchange_rate, billing_address, shipping_address, is_recurring, recurrence_pattern, next_invoice_date, pdf_url, document_hash, reminder_sent_count, last_reminder_sent_at, notes, internal_notes, reference_number, created_by_user_id, metadata, tags, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: kiosk_sessions; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.kiosk_sessions (id, pos_terminal_id, store_id, session_token, status, opened_at, closed_at, metadata, created_at) FROM stdin;
\.


--
-- Data for Name: loyalty_redemption_rules; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.loyalty_redemption_rules (id, organization_id, rule_name, points_earning_rate, points_redemption_rate, min_points_to_redeem, max_points_per_txn, max_redemption_percent, eligible_product_types, expiry_days, is_active, valid_from, valid_to, metadata, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: menu_categories; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.menu_categories (id, store_id, parent_category_id, name, code, description, category_level, display_order, icon, image_url, is_active, metadata, created_at, updated_at) FROM stdin;
9	8	\N	Breakfast	CAT-BRK	Morning delights	1	1	\N	\N	t	{}	2026-07-18 08:04:17.41459	2026-07-18 08:04:17.41459
10	8	\N	Appetizers	CAT-APP	Small bites and starters	1	2	\N	\N	t	{}	2026-07-18 08:04:17.41459	2026-07-18 08:04:17.41459
11	8	\N	Salads	CAT-SAL	Fresh garden salads	1	3	\N	\N	t	{}	2026-07-18 08:04:17.41459	2026-07-18 08:04:17.41459
12	8	\N	Main Courses	CAT-MAIN	Hearty meals and grills	1	4	\N	\N	t	{}	2026-07-18 08:04:17.41459	2026-07-18 08:04:17.41459
13	8	\N	Desserts	CAT-DES	Sweet treats	1	5	\N	\N	t	{}	2026-07-18 08:04:17.41459	2026-07-18 08:04:17.41459
14	8	\N	Hot Beverages	CAT-BEV-H	Coffee and tea	1	6	\N	\N	t	{}	2026-07-18 08:04:17.41459	2026-07-18 08:04:17.41459
15	8	\N	Cold Beverages	CAT-BEV-C	Smoothies and soft drinks	1	7	\N	\N	t	{}	2026-07-18 08:04:17.41459	2026-07-18 08:04:17.41459
\.


--
-- Data for Name: menu_item_availability_schedules; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.menu_item_availability_schedules (id, menu_item_id, day_of_week, start_time, end_time, is_active, valid_from, valid_to, metadata, created_at) FROM stdin;
\.


--
-- Data for Name: menu_item_modifiers; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.menu_item_modifiers (id, menu_item_id, modifier_name, modifier_type, price_adjustment, is_active, display_order, metadata, created_at) FROM stdin;
\.


--
-- Data for Name: menu_items; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.menu_items (id, store_id, menu_category_id, product_id, recipe_id, name, short_name, description, image_url, base_price, cost_price, preparation_time_min, tax_category_id, is_available, is_active, display_order, metadata, created_at, updated_at) FROM stdin;
5	8	14	\N	3	Double Espresso	Espresso	Premium Arabica double shot	\N	12.00	2.50	2	1	t	t	1	{}	2026-07-18 08:04:17.41459	2026-07-18 08:04:17.41459
6	8	9	\N	4	Classic Club Sandwich	Club Sandwich	Toasted triple-decker sandwich	\N	35.00	12.00	15	1	t	t	2	{}	2026-07-18 08:04:17.41459	2026-07-18 08:04:17.41459
7	8	15	\N	\N	Fresh Orange Juice	Orange Juice	100% freshly squeezed	\N	18.00	5.00	5	1	t	t	3	{}	2026-07-18 08:04:17.41459	2026-07-18 08:04:17.41459
8	8	12	\N	\N	Grilled Chicken Breast	Grilled Chicken	Marinated chicken with sides	\N	45.00	15.00	20	1	t	t	4	{}	2026-07-18 08:04:17.41459	2026-07-18 08:04:17.41459
10	8	9	11	5	Classic Eggs Benedict	Eggs Benedict	Two perfectly poached eggs with premium Canadian bacon on a toasted English muffin, topped with silky warm Hollandaise sauce and a sprinkle of fresh chives. Served with hash browns.	\N	10.95	\N	15	1	t	t	0	{}	2026-08-03 06:31:37.414242	2026-08-03 06:31:37.414242
\.


--
-- Data for Name: menu_modifier_groups; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.menu_modifier_groups (id, store_id, name, code, selection_type, min_selections, max_selections, is_active, display_order, metadata, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: menu_permissions; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.menu_permissions (id, menu_id, permission_id, metadata) FROM stdin;
1	2	1	{}
2	1	1	{}
3	2	2	{}
4	1	2	{}
5	2	3	{}
6	1	3	{}
7	3	4	{}
8	3	5	{}
9	3	6	{}
10	3	7	{}
11	4	8	{}
12	4	9	{}
13	4	10	{}
14	6	11	{}
15	5	11	{}
16	6	12	{}
17	5	12	{}
18	6	13	{}
19	5	13	{}
20	6	14	{}
21	5	14	{}
22	6	15	{}
23	5	15	{}
24	6	16	{}
25	5	16	{}
26	6	17	{}
27	5	17	{}
28	6	18	{}
29	5	18	{}
30	8	25	{}
31	7	25	{}
32	8	26	{}
33	7	26	{}
34	8	27	{}
35	7	27	{}
36	8	28	{}
37	7	28	{}
38	11	29	{}
39	10	29	{}
40	9	29	{}
41	11	30	{}
42	10	30	{}
43	9	30	{}
44	11	31	{}
45	10	31	{}
46	9	31	{}
47	11	32	{}
48	10	32	{}
49	9	32	{}
50	11	33	{}
51	10	33	{}
52	9	33	{}
53	11	34	{}
54	10	34	{}
55	9	34	{}
56	11	35	{}
57	10	35	{}
58	9	35	{}
59	13	36	{}
60	12	36	{}
61	13	37	{}
62	12	37	{}
63	13	38	{}
64	12	38	{}
65	13	39	{}
66	12	39	{}
67	13	40	{}
68	12	40	{}
69	16	41	{}
71	14	41	{}
72	16	42	{}
74	14	42	{}
75	16	43	{}
77	14	43	{}
78	16	44	{}
80	14	44	{}
81	16	45	{}
83	14	45	{}
84	20	46	{}
85	19	46	{}
86	18	46	{}
87	17	46	{}
88	20	47	{}
89	19	47	{}
90	18	47	{}
91	17	47	{}
92	20	48	{}
93	19	48	{}
94	18	48	{}
95	17	48	{}
96	20	49	{}
97	19	49	{}
98	18	49	{}
99	17	49	{}
100	20	50	{}
101	19	50	{}
102	18	50	{}
103	17	50	{}
104	21	51	{}
105	21	52	{}
106	21	53	{}
107	21	54	{}
108	22	55	{}
109	22	56	{}
110	22	57	{}
111	23	58	{}
112	23	59	{}
113	23	60	{}
114	23	61	{}
115	24	62	{}
116	24	63	{}
117	24	64	{}
118	24	65	{}
119	28	66	{}
120	27	66	{}
121	26	66	{}
122	25	66	{}
123	28	67	{}
124	27	67	{}
125	26	67	{}
126	25	67	{}
127	28	68	{}
128	27	68	{}
129	26	68	{}
130	25	68	{}
131	28	69	{}
132	27	69	{}
133	26	69	{}
134	25	69	{}
135	28	70	{}
136	27	70	{}
137	26	70	{}
138	25	70	{}
139	31	19	{}
140	30	19	{}
141	29	19	{}
142	31	20	{}
143	30	20	{}
144	29	20	{}
145	31	21	{}
146	30	21	{}
147	29	21	{}
148	31	22	{}
149	30	22	{}
150	29	22	{}
151	31	23	{}
152	30	23	{}
153	29	23	{}
154	31	24	{}
155	30	24	{}
156	29	24	{}
157	31	71	{}
158	30	71	{}
159	29	71	{}
160	31	72	{}
161	30	72	{}
162	29	72	{}
163	31	73	{}
164	30	73	{}
165	29	73	{}
171	14	97	{}
178	14	103	{}
179	14	104	{}
183	48	101	{}
184	48	100	{}
185	48	102	{}
186	49	101	{}
187	49	100	{}
188	49	102	{}
189	50	90	{}
190	50	89	{}
191	51	89	{}
192	51	90	{}
193	52	89	{}
194	53	122	{}
195	53	123	{}
196	53	124	{}
197	53	125	{}
198	53	126	{}
199	53	127	{}
200	53	128	{}
201	53	129	{}
202	53	130	{}
203	53	131	{}
204	53	132	{}
205	53	133	{}
206	53	134	{}
207	53	135	{}
208	53	136	{}
209	53	137	{}
210	53	138	{}
211	53	139	{}
212	53	141	{}
213	53	142	{}
214	53	143	{}
215	53	144	{}
216	54	145	{}
217	54	146	{}
218	54	147	{}
219	55	148	{}
220	55	149	{}
221	55	150	{}
222	55	151	{}
\.


--
-- Data for Name: menus; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.menus (id, module_id, parent_menu_id, name, code, route_path, icon, display_order, is_active, metadata, created_at, updated_at) FROM stdin;
1	1	\N	Overview	overview	/dashboard/overview	home	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
2	1	\N	Analytics	analytics	/dashboard/analytics	trending-up	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
3	2	\N	Tenants	tenants	/admin/tenants	building	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
4	3	\N	Organizations	organizations	/admin/organizations	briefcase	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
5	4	\N	Users	users	/admin/users	users	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
6	4	\N	Roles & Permissions	roles_permissions	/admin/roles	shield	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
7	5	\N	Stores	stores	/stores/list	store	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
8	5	\N	Storage Locations	storage_locations	/stores/locations	map-pin	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
9	6	\N	POS Transactions	pos_transactions	/pos/transactions	credit-card	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
10	6	\N	POS Terminals	pos_terminals	/pos/terminals	monitor	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
11	6	\N	POS Reports	pos_reports	/pos/reports	file-text	3	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
12	7	\N	Cashiers	cashiers	/cashiers/list	user-check	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
13	7	\N	Cashier Sessions	cashier_sessions	/cashiers/sessions	clock	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
14	8	\N	Stock Overview	stock_overview	/inventory/overview	package	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
16	8	\N	Stock Counts	stock_counts	/inventory/counts	clipboard	3	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
17	9	\N	Products	products	/products/list	box	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
18	9	\N	Categories	categories	/products/categories	grid	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
19	9	\N	Brands	brands	/products/brands	tag	3	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
20	9	\N	Price Lists	price_lists	/products/price-lists	dollar-sign	4	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
21	10	\N	Customers	customers	/customers/list	user-circle	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
22	11	\N	Suppliers	suppliers	/suppliers/list	truck	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
23	12	\N	Purchase Orders	purchase_orders	/purchase-orders/list	file-text	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
24	13	\N	Sales Orders	sales_orders	/sales-orders/list	shopping-bag	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
25	14	\N	Sales Reports	sales_reports	/reports/sales	bar-chart	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
26	14	\N	Purchase Reports	purchase_reports	/reports/purchases	file-text	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
27	14	\N	Inventory Reports	inventory_reports	/reports/inventory	package	3	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
28	14	\N	Financial Reports	financial_reports	/reports/financial	dollar-sign	4	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
29	15	\N	UI Modules	ui_modules	/admin/ui-modules	layout	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
30	15	\N	System Settings	system_settings	/admin/settings	settings	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
31	15	\N	Audit Logs	audit_logs	/admin/audit-logs	file-text	3	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
48	20	\N	Promotions	promo_1	promotions-discount/promotion	trending	1	t	"{\\"color\\":\\"blue\\"}"	2026-07-31 05:55:19.231041	2026-07-31 05:55:19.231041
49	20	\N	Coupon	coupon_01	promotions-discount/coupon	money	2	t	"{\\"color\\":\\"blue\\"}"	2026-07-31 05:56:25.368484	2026-07-31 05:56:25.368484
50	21	\N	Catalog & Recipes	restaurant_catalog	restaurant/catalog	warehouse	1	t	"{\\"color\\":\\"blue\\"}"	2026-07-31 12:20:26.161617	2026-07-31 12:20:26.161617
51	21	\N	Dining Operations	dining_ops	restaurant/dining	dining-menu	2	t	"{\\"color\\":\\"blue\\"}"	2026-07-31 12:36:26.588971	2026-07-31 12:36:26.588971
52	21	\N	Kitchen & Quality	kitchen_mgmt	restaurant/kitchen	archive	3	t	"{\\"color\\":\\"blue\\"}"	2026-08-01 08:09:31.257378	2026-08-01 08:09:31.257378
53	22	\N	POS Operations	pos-operations	/pos/dashboard	cart	1	t	"{\\"color\\":\\"blue\\"}"	2026-08-11 06:48:48.068079	2026-08-11 06:48:48.068079
54	22	\N	POS Customers	pos-customers	/pos/dashboard	user-group	1	t	"{\\"color\\":\\"blue\\"}"	2026-08-11 06:49:43.396938	2026-08-11 06:49:43.396938
55	22	\N	POS Cashier Sessions	pos-c-s	/pos/dashboard	database	1	t	"{\\"color\\":\\"blue\\"}"	2026-08-11 06:50:32.458699	2026-08-11 06:50:32.458699
\.


--
-- Data for Name: module_permissions; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.module_permissions (id, module_id, permission_id, metadata) FROM stdin;
1	1	1	{}
2	1	2	{}
3	1	3	{}
4	2	4	{}
5	2	5	{}
6	2	6	{}
7	2	7	{}
8	3	8	{}
9	3	9	{}
10	3	10	{}
11	4	11	{}
12	4	12	{}
13	4	13	{}
14	4	14	{}
15	4	15	{}
16	4	16	{}
17	4	17	{}
18	4	18	{}
19	5	25	{}
20	5	26	{}
21	5	27	{}
22	5	28	{}
23	6	29	{}
24	6	30	{}
25	6	31	{}
26	6	32	{}
27	6	33	{}
28	6	34	{}
29	6	35	{}
30	7	36	{}
31	7	37	{}
32	7	38	{}
33	7	39	{}
34	7	40	{}
35	8	41	{}
36	8	42	{}
37	8	43	{}
38	8	44	{}
39	8	45	{}
40	9	46	{}
41	9	47	{}
42	9	48	{}
43	9	49	{}
44	9	50	{}
45	10	51	{}
46	10	52	{}
47	10	53	{}
48	10	54	{}
49	11	55	{}
50	11	56	{}
51	11	57	{}
52	12	58	{}
53	12	59	{}
54	12	60	{}
55	12	61	{}
56	13	62	{}
57	13	63	{}
58	13	64	{}
59	13	65	{}
60	14	66	{}
61	14	67	{}
62	14	68	{}
63	14	69	{}
64	14	70	{}
65	15	19	{}
66	15	20	{}
67	15	21	{}
68	15	22	{}
69	15	23	{}
70	15	24	{}
71	15	71	{}
72	15	72	{}
73	15	73	{}
81	4	96	{}
82	8	97	{}
88	8	103	{}
89	8	104	{}
95	20	110	{}
96	20	111	{}
97	20	112	{}
98	20	113	{}
99	20	114	{}
100	21	115	{}
101	21	116	{}
102	21	117	{}
103	21	118	{}
104	21	119	{}
105	21	120	{}
106	22	121	{}
107	22	122	{}
108	22	123	{}
109	22	124	{}
110	22	125	{}
111	22	126	{}
112	22	127	{}
113	22	128	{}
114	22	129	{}
115	22	130	{}
116	22	131	{}
117	22	132	{}
118	22	133	{}
119	22	134	{}
120	22	135	{}
121	22	136	{}
122	22	137	{}
123	22	138	{}
124	22	139	{}
125	22	141	{}
126	22	142	{}
127	22	143	{}
128	22	144	{}
129	22	145	{}
130	22	146	{}
131	22	147	{}
132	22	148	{}
133	22	149	{}
134	22	150	{}
135	22	151	{}
\.


--
-- Data for Name: modules; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.modules (id, name, code, description, icon, is_active, display_order, metadata, created_at, updated_at) FROM stdin;
1	Dashboard	dashboard	Main dashboard and overview with analytics	dashboard	t	1	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
2	Tenant Management	tenants	Multi-tenant configuration and management	building	t	2	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
3	Organization Setup	organizations	Organization and company structure management	briefcase	t	3	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
4	User Management	users	User accounts and authentication	users	t	4	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
5	Store Management	stores	Store locations and configuration	store	t	5	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
6	Point of Sale	pos	POS transactions and terminal management	shopping-cart	t	6	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
7	Cashier Operations	cashiers	Cashier management and session control	user-check	t	7	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
8	Inventory Management	inventory	Stock control and warehouse management	package	t	8	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
9	Product Catalog	products	Product master data and catalog	box	t	9	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
10	Customer Management	customers	Customer database and relationship management	user-circle	t	10	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
11	Supplier Management	suppliers	Supplier database and procurement	truck	t	11	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
12	Purchase Orders	purchase_orders	Purchase order processing	file-text	t	12	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
13	Sales Orders	sales_orders	Sales order management	shopping-bag	t	13	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
14	Reports & Analytics	reports	Business intelligence and reporting	bar-chart	t	14	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
15	System Administration	admin	System settings and configuration	settings	t	15	{}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
20	Promotions and Discounts	PROMO_DISCOUNT	Module for Creating  and managing the promotions and discounts	inventory_2	t	16	{}	2026-07-31 05:53:54.310511	2026-07-31 05:53:54.310511
21	Restaurant Management	RESTAURANT	For managing the restaurant	corporate_fare	t	17	{}	2026-07-31 12:14:49.263067	2026-07-31 12:14:49.263067
22	Sale	SALE	Unified sales module containing Retail POS, Wholesale, Restaurant POS, transactions, stock, and sessions.	shopping_cart	t	18	{}	2026-08-11 06:45:08.560209	2026-08-11 06:45:08.560209
\.


--
-- Data for Name: order_fulfillment_items; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.order_fulfillment_items (id, fulfillment_id, order_line_id, organization_id, quantity_fulfilled, batch_number, serial_numbers, created_at) FROM stdin;
\.


--
-- Data for Name: order_fulfillments; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.order_fulfillments (id, sales_order_id, organization_id, fulfillment_number, fulfillment_status, shipment_status, fulfillment_store_id, shipping_carrier, shipping_method, tracking_number, tracking_url, picked_at, packed_at, shipped_at, estimated_delivery_date, actual_delivery_date, picked_by_user_id, packed_by_user_id, notes, metadata, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: order_status_history; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.order_status_history (id, sales_order_id, organization_id, from_status, to_status, reason, notes, changed_by_user_id, changed_at) FROM stdin;
\.


--
-- Data for Name: organizations; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.organizations (id, name, code, legal_name, tax_id, currency_code, fiscal_year_variant, is_active, metadata, created_at, updated_at) FROM stdin;
1	Qitaf Group	ORG001	Qitaf Group LLC	TAX123456789	SAR	CALENDAR	t	{"industry": "retail", "established": "2020"}	2026-07-18 07:58:11.381095	2026-07-18 07:58:11.381095
\.


--
-- Data for Name: permissions; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.permissions (id, name, code, description, metadata, created_at) FROM stdin;
1	View Dashboard	dashboard:view	Can view dashboard and analytics	{}	2026-07-18 07:59:03.504911
2	Manage Dashboard	dashboard:manage	Can customize and manage dashboard layouts	{}	2026-07-18 07:59:03.504911
3	Export Dashboard	dashboard:export	Can export dashboard reports	{}	2026-07-18 07:59:03.504911
4	View Tenants	tenants:view	Can view tenant information	{}	2026-07-18 07:59:03.504911
5	Manage Tenants	tenants:manage	Can create and edit tenants	{}	2026-07-18 07:59:03.504911
6	Delete Tenants	tenants:delete	Can delete tenants	{}	2026-07-18 07:59:03.504911
7	Configure Tenants	tenants:configure	Can configure tenant settings and features	{}	2026-07-18 07:59:03.504911
8	View Organizations	organizations:view	Can view organization details	{}	2026-07-18 07:59:03.504911
9	Manage Organizations	organizations:manage	Can create and edit organizations	{}	2026-07-18 07:59:03.504911
10	Delete Organizations	organizations:delete	Can delete organizations	{}	2026-07-18 07:59:03.504911
11	View Users	users:view	Can view user list and details	{}	2026-07-18 07:59:03.504911
12	Manage Users	users:manage	Can create and edit users	{}	2026-07-18 07:59:03.504911
13	Delete Users	users:delete	Can delete users	{}	2026-07-18 07:59:03.504911
14	Reset User Password	users:reset_password	Can reset user passwords	{}	2026-07-18 07:59:03.504911
15	View Roles	roles:view	Can view roles list	{}	2026-07-18 07:59:03.504911
16	Manage Roles	roles:manage	Can create and edit roles	{}	2026-07-18 07:59:03.504911
17	Delete Roles	roles:delete	Can delete roles	{}	2026-07-18 07:59:03.504911
18	Assign Roles	roles:assign	Can assign roles to users	{}	2026-07-18 07:59:03.504911
19	View UI Modules	ui_modules:view	Can view UI modules and menus	{}	2026-07-18 07:59:03.504911
20	Manage UI Modules	ui_modules:manage	Can create and edit UI modules	{}	2026-07-18 07:59:03.504911
21	Delete UI Modules	ui_modules:delete	Can delete UI modules	{}	2026-07-18 07:59:03.504911
22	View Permissions	permissions:view	Can view permissions	{}	2026-07-18 07:59:03.504911
23	Manage Permissions	permissions:manage	Can create and edit permissions	{}	2026-07-18 07:59:03.504911
24	Delete Permissions	permissions:delete	Can delete permissions	{}	2026-07-18 07:59:03.504911
25	View Stores	stores:view	Can view store information	{}	2026-07-18 07:59:03.504911
26	Manage Stores	stores:manage	Can create and edit stores	{}	2026-07-18 07:59:03.504911
27	Delete Stores	stores:delete	Can delete stores	{}	2026-07-18 07:59:03.504911
28	Configure Stores	stores:configure	Can configure store settings	{}	2026-07-18 07:59:03.504911
29	View POS	pos:view	Can view POS transactions and terminals	{}	2026-07-18 07:59:03.504911
30	Manage POS	pos:manage	Can configure POS terminals and settings	{}	2026-07-18 07:59:03.504911
31	Process Sales	pos:process_sales	Can process sales transactions	{}	2026-07-18 07:59:03.504911
32	Void Transactions	pos:void_transactions	Can void POS transactions	{}	2026-07-18 07:59:03.504911
33	Apply Discounts	pos:apply_discounts	Can apply discounts to transactions	{}	2026-07-18 07:59:03.504911
34	Process Returns	pos:process_returns	Can process return transactions	{}	2026-07-18 07:59:03.504911
35	View POS Reports	pos:view_reports	Can view POS reports and analytics	{}	2026-07-18 07:59:03.504911
36	View Cashiers	cashiers:view	Can view cashier information	{}	2026-07-18 07:59:03.504911
37	Manage Cashiers	cashiers:manage	Can create and edit cashiers	{}	2026-07-18 07:59:03.504911
38	Delete Cashiers	cashiers:delete	Can delete cashiers	{}	2026-07-18 07:59:03.504911
39	Manage Sessions	cashiers:manage_sessions	Can open/close cashier sessions	{}	2026-07-18 07:59:03.504911
40	View Sessions	cashiers:view_sessions	Can view cashier session history	{}	2026-07-18 07:59:03.504911
41	View Inventory	inventory:view	Can view inventory levels and stock	{}	2026-07-18 07:59:03.504911
42	Manage Inventory	inventory:manage	Can adjust inventory and manage stock	{}	2026-07-18 07:59:03.504911
43	Transfer Inventory	inventory:transfer	Can transfer stock between locations	{}	2026-07-18 07:59:03.504911
44	Conduct Stock Count	inventory:stock_count	Can perform stock counts	{}	2026-07-18 07:59:03.504911
45	View Inventory Reports	inventory:view_reports	Can view inventory reports	{}	2026-07-18 07:59:03.504911
46	View Products	products:view	Can view product catalog	{}	2026-07-18 07:59:03.504911
47	Manage Products	products:manage	Can create and edit products	{}	2026-07-18 07:59:03.504911
48	Delete Products	products:delete	Can delete products	{}	2026-07-18 07:59:03.504911
49	Manage Pricing	products:manage_pricing	Can manage product prices	{}	2026-07-18 07:59:03.504911
50	View Cost Prices	products:view_cost	Can view product cost prices	{}	2026-07-18 07:59:03.504911
51	View Customers	customers:view	Can view customer information	{}	2026-07-18 07:59:03.504911
52	Manage Customers	customers:manage	Can create and edit customers	{}	2026-07-18 07:59:03.504911
53	Delete Customers	customers:delete	Can delete customers	{}	2026-07-18 07:59:03.504911
54	View Customer History	customers:view_history	Can view customer purchase history	{}	2026-07-18 07:59:03.504911
55	View Suppliers	suppliers:view	Can view supplier information	{}	2026-07-18 07:59:03.504911
56	Manage Suppliers	suppliers:manage	Can create and edit suppliers	{}	2026-07-18 07:59:03.504911
57	Delete Suppliers	suppliers:delete	Can delete suppliers	{}	2026-07-18 07:59:03.504911
58	View Purchase Orders	purchase_orders:view	Can view purchase orders	{}	2026-07-18 07:59:03.504911
59	Manage Purchase Orders	purchase_orders:manage	Can create and edit purchase orders	{}	2026-07-18 07:59:03.504911
60	Delete Purchase Orders	purchase_orders:delete	Can delete purchase orders	{}	2026-07-18 07:59:03.504911
61	Approve Purchase Orders	purchase_orders:approve	Can approve purchase orders	{}	2026-07-18 07:59:03.504911
62	View Sales Orders	sales_orders:view	Can view sales orders	{}	2026-07-18 07:59:03.504911
63	Manage Sales Orders	sales_orders:manage	Can create and edit sales orders	{}	2026-07-18 07:59:03.504911
64	Delete Sales Orders	sales_orders:delete	Can delete sales orders	{}	2026-07-18 07:59:03.504911
65	Approve Sales Orders	sales_orders:approve	Can approve sales orders	{}	2026-07-18 07:59:03.504911
66	View Sales Reports	reports:sales	Can view sales reports and analytics	{}	2026-07-18 07:59:03.504911
67	View Purchase Reports	reports:purchases	Can view purchase reports	{}	2026-07-18 07:59:03.504911
68	View Inventory Reports	reports:inventory	Can view inventory reports	{}	2026-07-18 07:59:03.504911
69	View Financial Reports	reports:financial	Can view P&L and financial reports	{}	2026-07-18 07:59:03.504911
70	Export Reports	reports:export	Can export reports to various formats	{}	2026-07-18 07:59:03.504911
71	View Settings	settings:view	Can view system settings	{}	2026-07-18 07:59:03.504911
72	Manage Settings	settings:manage	Can modify system settings	{}	2026-07-18 07:59:03.504911
73	View Audit Logs	settings:audit_logs	Can view system audit logs	{}	2026-07-18 07:59:03.504911
88	View Restaurant	restaurant:view	Can view restaurant operations	{}	2026-07-18 08:04:42.348313
89	Manage Menu	restaurant:menu_manage	Can manage menu categories and items	{}	2026-07-18 08:04:42.348313
90	Manage Recipes	restaurant:recipe_manage	Can manage recipes and ingredients	{}	2026-07-18 08:04:42.348313
91	Manage Tables	restaurant:table_manage	Can manage restaurant floor plan and tables	{}	2026-07-18 08:04:42.348313
92	Process Orders	restaurant:process_orders	Can take and process restaurant orders	{}	2026-07-18 08:04:42.348313
93	View Kitchen Display	restaurant:kitchen_view	Can view and manage kitchen display	{}	2026-07-18 08:04:42.348313
94	Restaurant Settings	restaurant:settings	Can configure restaurant specific settings	{}	2026-07-18 08:04:42.348313
96	update	users:update	USer can delete	{}	2026-07-28 09:13:43.361288
98	Promotions & Discounts View	promotions_&_discounts:view	Can view Promotions & Discounts	{}	2026-07-29 11:10:45.773547
99	Promotions & Discounts Manage	promotions_&_discounts:manage	Can manage Promotions & Discounts	{}	2026-07-29 11:10:45.775213
100	Promotions & Discounts Delete	promotions_&_discounts:delete	Can delete Promotions & Discounts	{}	2026-07-29 11:10:45.776523
101	Promotions & Discounts Add	promotions_&_discounts:add	Can add Promotions & Discounts	{}	2026-07-29 11:10:45.777955
102	Promotions & Discounts List	promotions_&_discounts:list	Can list Promotions & Discounts	{}	2026-07-29 11:10:45.779248
103	shiped	inventory:shiped	Assign permission user can shipped the stock transfer flow	{}	2026-07-30 07:54:02.881955
104	receive	inventory:receive	Assign permission user can received the stock transfer flow	{}	2026-07-30 07:54:02.884484
105	permission module test View	permission_module_test:view	Can view permission module test	{}	2026-07-30 08:15:59.66262
106	permission module test Manage	permission_module_test:manage	Can manage permission module test	{}	2026-07-30 08:15:59.664649
107	permission module test Delete	permission_module_test:delete	Can delete permission module test	{}	2026-07-30 08:15:59.665892
108	permission module test Add	permission_module_test:add	Can add permission module test	{}	2026-07-30 08:15:59.66711
109	permission module test List	permission_module_test:list	Can list permission module test	{}	2026-07-30 08:15:59.668411
97	trnasfer	inventory:trnasfer	user allow to direct transfer of stock movement	{}	2026-07-29 06:28:03.380723
110	Promotions and Discounts View	promotions_and_discounts:view	Can view Promotions and Discounts	{}	2026-07-31 05:53:54.825639
111	Promotions and Discounts Manage	promotions_and_discounts:manage	Can manage Promotions and Discounts	{}	2026-07-31 05:53:54.827618
112	Promotions and Discounts Delete	promotions_and_discounts:delete	Can delete Promotions and Discounts	{}	2026-07-31 05:53:54.828985
113	Promotions and Discounts Add	promotions_and_discounts:add	Can add Promotions and Discounts	{}	2026-07-31 05:53:54.830387
114	Promotions and Discounts List	promotions_and_discounts:list	Can list Promotions and Discounts	{}	2026-07-31 05:53:54.831689
115	Restaurant Management View	restaurant_management:view	Can view Restaurant Management	{}	2026-07-31 12:14:49.798078
116	Restaurant Management Manage	restaurant_management:manage	Can manage Restaurant Management	{}	2026-07-31 12:14:49.799895
117	Restaurant Management Delete	restaurant_management:delete	Can delete Restaurant Management	{}	2026-07-31 12:14:49.801292
118	Restaurant Management Add	restaurant_management:add	Can add Restaurant Management	{}	2026-07-31 12:14:49.802811
119	Restaurant Management List	restaurant_management:list	Can list Restaurant Management	{}	2026-07-31 12:14:49.804229
120	Restaurant Management Configure	restaurant_management:configure	Can configure Restaurant Management	{}	2026-07-31 12:14:49.8056
121	Sale View	sale:view	Can view Sale	{}	2026-08-11 06:45:09.097408
122	view-dashboard	sale:view-dashboard	view-dashboard	{}	2026-08-11 07:15:55.539055
123	view products	sale:view_products	view items	{}	2026-08-11 07:17:20.70082
124	list view	sale:list_view	list	{}	2026-08-11 07:17:20.702219
125	grid view	sale:grid_view	grid view	{}	2026-08-11 07:17:20.703511
126	add to cart	sale:add_to_cart	add	{}	2026-08-11 07:21:10.430837
127	view product	sale:view_product	can view product details	{}	2026-08-11 07:21:10.432435
128	view retail pricelist	sale:view_retail_pricelist	can view retail price list	{}	2026-08-11 07:22:57.69737
129	view wholesale pricelist	sale:view_wholesale_pricelist	wholesale	{}	2026-08-11 07:23:23.770474
130	view restaurant mode	sale:view_restaurant_mode	restaurant mode	{}	2026-08-11 07:23:57.984951
131	view bills	sale:view_bills	view bills	{}	2026-08-11 07:25:39.846663
132	edit bills	sale:edit_bills	can edit bill	{}	2026-08-11 07:25:39.848543
133	void transaction	sale:void_transaction	can void bill	{}	2026-08-11 07:25:39.849854
134	Open cart	sale:open_cart	can open cart	{}	2026-08-11 07:34:37.730195
135	add items	sale:add_items	add items	{}	2026-08-11 07:34:37.731924
136	delete items	sale:delete_items	delete items	{}	2026-08-11 07:34:37.733178
137	draft order	sale:draft_order	can draft order	{}	2026-08-11 07:34:37.73469
138	biometric on delete	sale:biometric_on_delete	need biometric on deleting an item	{}	2026-08-11 07:34:37.735981
139	biometric draft order	sale:biometric_draft_order	need biometric when drafting order	{}	2026-08-11 07:34:37.737209
141	view collection	sale:view_collection	view	{}	2026-08-11 07:36:03.948654
142	submit collection	sale:submit_collection	can submit collection	{}	2026-08-11 07:36:03.950286
143	view stock levels	sale:view_stock_levels	can view stock levels	{}	2026-08-11 07:37:22.95237
144	request stock	sale:request_stock	stock	{}	2026-08-11 07:37:22.953815
145	view list	sale:view_list	view list of customers	{}	2026-08-11 07:48:08.706109
146	add customer	sale:add_customer	add customer	{}	2026-08-11 07:48:48.492532
147	add guest	sale:add_guest	guest	{}	2026-08-11 07:48:48.494029
148	open session	sale:open_session	can open session	{}	2026-08-11 07:49:23.628172
149	close session	sale:close_session	can close session	{}	2026-08-11 07:50:05.180746
150	view session	sale:view_session	view current session	{}	2026-08-11 07:50:47.173252
151	view summary	sale:view_summary	view session summary	{}	2026-08-11 07:53:13.447881
\.


--
-- Data for Name: pos_payments; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.pos_payments (id, transaction_id, payment_method, payment_gateway, amount, payment_reference, reference_number, payment_date, metadata, created_at) FROM stdin;
1	1	cash	\N	216.78	\N	\N	2026-07-20 08:21:31.565602	{}	2026-07-20 08:21:31.565602
2	2	cash	\N	375.41	\N	\N	2026-07-20 08:31:16.428288	{}	2026-07-20 08:31:16.428288
3	3	cash	\N	48.75	\N	\N	2026-07-20 11:03:02.748107	{}	2026-07-20 11:03:02.748107
4	4	cash	\N	32.72	\N	\N	2026-07-20 11:43:31.240882	{}	2026-07-20 11:43:31.240882
5	5	cash	\N	39.03	\N	\N	2026-07-21 07:11:31.136798	{}	2026-07-21 07:11:31.136798
6	6	cash	\N	48.75	\N	\N	2026-07-21 07:12:23.919513	{}	2026-07-21 07:12:23.919513
7	7	cash	\N	348.97	\N	\N	2026-07-21 07:12:48.88868	{}	2026-07-21 07:12:48.88868
8	8	cash	\N	24.09	\N	\N	2026-07-22 06:26:08.081026	{}	2026-07-22 06:26:08.081026
9	9	cash	\N	43.58	\N	\N	2026-07-22 07:37:06.232501	{}	2026-07-22 07:37:06.232501
10	10	cash	\N	334.65	\N	\N	2026-07-22 07:37:39.286682	{}	2026-07-22 07:37:39.286682
11	11	cash	\N	488.75	\N	\N	2026-07-22 07:38:22.797239	{}	2026-07-22 07:38:22.797239
12	12	cash	\N	645.10	\N	\N	2026-07-22 09:05:40.947528	{}	2026-07-22 09:05:40.947528
13	13	cash	\N	282.35	\N	\N	2026-07-22 09:52:19.508288	{}	2026-07-22 09:52:19.508288
14	14	cash	\N	286.38	\N	\N	2026-07-22 09:54:51.091945	{}	2026-07-22 09:54:51.091945
15	15	cash	\N	25.23	\N	\N	2026-07-22 10:01:20.581731	{}	2026-07-22 10:01:20.581731
16	16	cash	\N	10.34	\N	\N	2026-07-22 10:02:26.804268	{}	2026-07-22 10:02:26.804268
17	17	cash	\N	5.17	\N	\N	2026-07-22 10:02:51.759676	{}	2026-07-22 10:02:51.759676
18	18	cash	\N	5.17	\N	\N	2026-07-22 10:05:41.936458	{}	2026-07-22 10:05:41.936458
19	19	cash	\N	5.17	\N	\N	2026-07-22 10:07:14.543004	{}	2026-07-22 10:07:14.543004
20	20	cash	\N	393.52	\N	\N	2026-07-22 13:11:05.684736	{}	2026-07-22 13:11:05.684736
21	21	cash	\N	207.00	\N	\N	2026-07-22 13:19:08.054936	{}	2026-07-22 13:19:08.054936
22	22	cash	\N	47.03	\N	\N	2026-07-23 06:18:29.351493	{}	2026-07-23 06:18:29.351493
23	23	cash	\N	37.89	\N	\N	2026-07-23 06:28:34.511959	{}	2026-07-23 06:28:34.511959
24	24	cash	\N	5.17	\N	\N	2026-07-23 06:37:55.647437	{}	2026-07-23 06:37:55.647437
25	25	cash	\N	19.49	\N	\N	2026-07-23 08:31:04.335279	{}	2026-07-23 08:31:04.335279
26	26	cash	\N	5.17	\N	\N	2026-07-23 10:54:04.989587	{}	2026-07-23 10:54:04.989587
27	27	cash	\N	596.15	\N	\N	2026-07-23 11:07:35.581775	{}	2026-07-23 11:07:35.581775
28	28	cash	\N	256.95	\N	\N	2026-07-23 11:16:48.175871	{}	2026-07-23 11:16:48.175871
29	29	cash	\N	37.83	\N	\N	2026-07-23 11:24:59.701592	{}	2026-07-23 11:24:59.701592
30	30	cash	\N	5.17	\N	\N	2026-07-23 11:42:46.173833	{"customer_name": "Guest ضيف"}	2026-07-23 11:42:46.173833
31	31	cash	\N	70.77	\N	\N	2026-07-24 09:14:58.950928	{"customer_name": "Guest ضيف"}	2026-07-24 09:14:58.950928
32	32	cash	\N	28.69	\N	\N	2026-07-27 10:47:45.09573	{"customer_name": "Guest ضيف"}	2026-07-27 10:47:45.09573
33	33	cash	\N	33.86	\N	\N	2026-07-27 11:06:26.270277	{"customer_name": "Guest ضيف"}	2026-07-27 11:06:26.270277
34	34	cash	\N	28.69	\N	\N	2026-07-27 11:11:19.937977	{"customer_name": "Guest ضيف"}	2026-07-27 11:11:19.937977
35	35	cash	\N	65.52	\N	\N	2026-07-27 11:42:20.385587	{"customer_name": "Guest ضيف"}	2026-07-27 11:42:20.385587
36	36	cash	\N	309.91	\N	\N	2026-07-27 11:56:46.184484	{"customer_name": "Guest ضيف"}	2026-07-27 11:56:46.184484
37	37	cash	\N	832.61	\N	\N	2026-07-27 12:05:02.760443	{"customer_name": "Guest ضيف"}	2026-07-27 12:05:02.760443
38	38	cash	\N	137.65	\N	\N	2026-07-27 12:06:45.4173	{"customer_name": "Guest ضيف"}	2026-07-27 12:06:45.4173
39	39	cash	\N	34.38	\N	\N	2026-07-28 08:35:54.536829	{"customer_name": "Guest ضيف"}	2026-07-28 08:35:54.536829
40	40	cash	\N	48.81	\N	\N	2026-07-28 08:37:14.683579	{"customer_name": "Guest ضيف"}	2026-07-28 08:37:14.683579
41	41	cash	\N	130.75	\N	\N	2026-07-30 08:02:53.798716	{"customer_name": "Guest ضيف"}	2026-07-30 08:02:53.798716
42	42	cash	\N	33.86	\N	\N	2026-07-30 08:10:35.389189	{"customer_name": "Guest ضيف"}	2026-07-30 08:10:35.389189
43	43	cash	\N	91.24	\N	\N	2026-07-30 10:59:28.638998	{"customer_name": "Guest ضيف"}	2026-07-30 10:59:28.638998
44	44	cash	\N	630.62	\N	\N	2026-07-30 11:04:16.982987	{"customer_name": "Guest ضيف"}	2026-07-30 11:04:16.982987
45	45	cash	\N	582.27	\N	\N	2026-07-30 11:08:54.635742	{"customer_name": "Guest ضيف"}	2026-07-30 11:08:54.635742
46	46	cash	\N	66.73	\N	\N	2026-07-31 09:02:44.084626	{"customer_name": "Guest ضيف"}	2026-07-31 09:02:44.084626
47	47	cash	\N	4.03	\N	\N	2026-07-31 09:05:05.133156	{"customer_name": "Guest ضيف"}	2026-07-31 09:05:05.133156
48	48	cash	\N	73.47	\N	\N	2026-07-31 09:11:02.499558	{"customer_name": "Guest ضيف"}	2026-07-31 09:11:02.499558
49	49	cash	\N	33.86	\N	\N	2026-07-31 09:13:36.25372	{"customer_name": "Guest ضيف"}	2026-07-31 09:13:36.25372
50	50	cash	\N	228.94	\N	\N	2026-07-31 10:19:10.768834	{"customer_name": "Guest ضيف"}	2026-07-31 10:19:10.768834
51	51	cash	\N	46.52	\N	\N	2026-08-01 06:18:24.480989	{"customer_name": "Guest ضيف"}	2026-08-01 06:18:24.480989
52	52	cash	\N	16.12	\N	\N	2026-08-01 06:26:48.177285	{"customer_name": "Guest ضيف"}	2026-08-01 06:26:48.177285
53	53	cash	\N	32.72	\N	\N	2026-08-01 06:47:31.614154	{"customer_name": "Guest ضيف"}	2026-08-01 06:47:31.614154
54	54	cash	\N	411.89	\N	\N	2026-08-03 05:31:22.85154	{"customer_name": "Guest ضيف"}	2026-08-03 05:31:22.85154
55	55	cash	\N	65.37	\N	\N	2026-08-03 09:55:00.691199	{"customer_name": "Walk-in Guest"}	2026-08-03 09:55:00.691199
56	56	cash	\N	5.17	\N	\N	2026-08-03 10:34:21.544893	{"customer_name": "Guest ضيف"}	2026-08-03 10:34:21.544893
57	57	cash	\N	5.17	\N	\N	2026-08-03 10:36:05.380651	{"customer_name": "Guest ضيف"}	2026-08-03 10:36:05.380651
58	58	cash	\N	9.20	\N	\N	2026-08-03 10:36:47.67446	{"customer_name": "Guest ضيف"}	2026-08-03 10:36:47.67446
59	59	cash	\N	20.70	\N	\N	2026-08-03 11:39:23.359173	{"customer_name": "Guest ضيف"}	2026-08-03 11:39:23.359173
60	60	cash	\N	4.03	\N	\N	2026-08-03 11:41:39.530727	{"customer_name": "Guest ضيف"}	2026-08-03 11:41:39.530727
61	61	cash	\N	8.06	\N	\N	2026-08-11 11:24:31.159881	{"customer_name": "Guest ضيف"}	2026-08-11 11:24:31.159881
62	62	cash	\N	47.70	\N	\N	2026-08-11 11:46:45.44484	{"customer_name": "Guest ضيف"}	2026-08-11 11:46:45.44484
\.


--
-- Data for Name: pos_terminals; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.pos_terminals (id, store_id, terminal_code, terminal_name, device_id, is_active, metadata, created_at, updated_at) FROM stdin;
1	1	POS-RYD-01	Checkout Counter 1	DEVICE-RYD-001	t	{"location": "Front"}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
2	1	POS-RYD-02	Checkout Counter 2	DEVICE-RYD-002	t	{"location": "Front"}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
3	1	POS-RYD-03	Checkout Counter 3	DEVICE-RYD-003	t	{"location": "Front"}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
4	1	POS-RYD-04	Express Checkout	DEVICE-RYD-004	t	{"location": "Express Lane"}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
5	2	POS-JED-01	Checkout Counter 1	DEVICE-JED-001	t	{"location": "Front"}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
6	2	POS-JED-02	Checkout Counter 2	DEVICE-JED-002	t	{"location": "Front"}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
7	5	POS-WHSL-01	Wholesale Counter	DEVICE-WHSL-001	t	{"location": "Main Counter"}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
8	1	POS-001	Main Counter Terminal 1	DEVICE-MAIN-001	t	{}	2026-07-18 07:59:38.038245	2026-07-18 07:59:38.038245
9	1	POS-002	Main Counter Terminal 2	DEVICE-MAIN-002	t	{}	2026-07-18 07:59:38.038245	2026-07-18 07:59:38.038245
12	8	001	test	dev-0001	t	\N	2026-07-21 06:33:40.727823	2026-07-21 06:33:40.727823
14	9	002	test2	dev-0002	t	\N	2026-07-21 10:10:51.121019	2026-07-21 10:10:51.121019
13	9	001	testt	dev-0001	t	\N	2026-07-21 09:51:34.005688	2026-07-21 10:53:31.287372
10	2	POS-003	\N	\N	t	\N	2026-07-18 07:59:38.038245	2026-07-24 10:43:20.390897
15	9	top-01	Test 01	ckjzkc88	t	\N	2026-07-24 10:46:31.09361	2026-07-24 10:46:31.09361
17	10	Nast-001	Corner Counter	3216468	t	\N	2026-07-30 05:31:42.984262	2026-07-30 05:31:42.984262
18	17	pos002	counter2	dev128	t	\N	2026-07-30 11:42:55.354523	2026-07-30 11:42:55.354523
\.


--
-- Data for Name: pos_transaction_lines; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.pos_transaction_lines (id, transaction_id, product_id, product_variant_id, quantity, uom_id, unit_price, discount_amount, tax_amount, subtotal, line_total, cost_price, line_number, serial_number, batch_number, metadata, created_at) FROM stdin;
1	1	8	\N	10.000	\N	18.0000	0.00	0.00	180.00	207.00	0.00	\N	\N	\N	{}	2026-07-20 08:21:31.561314
2	1	2	\N	1.000	\N	8.5000	0.00	0.00	8.50	9.78	0.00	\N	\N	\N	{}	2026-07-20 08:21:31.561314
3	2	40	\N	6.000	\N	48.5000	0.00	0.00	291.00	334.65	0.00	\N	\N	\N	{}	2026-07-20 08:31:16.424699
4	2	23	\N	1.000	\N	19.9500	0.00	0.00	19.95	22.94	0.00	\N	\N	\N	{}	2026-07-20 08:31:16.424699
5	2	7	\N	1.000	\N	15.5000	0.00	0.00	15.50	17.82	0.00	\N	\N	\N	{}	2026-07-20 08:31:16.424699
6	3	34	\N	1.000	\N	24.9500	0.00	0.00	24.95	28.69	0.00	\N	\N	\N	{}	2026-07-20 11:03:02.744034
7	3	26	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-20 11:03:02.744034
8	3	6	\N	1.000	\N	12.9500	0.00	0.00	12.95	14.89	0.00	\N	\N	\N	{}	2026-07-20 11:03:02.744034
9	4	34	\N	1.000	\N	24.9500	0.00	0.00	24.95	28.69	0.00	\N	\N	\N	{}	2026-07-20 11:43:31.238534
10	4	35	\N	1.000	\N	3.5000	0.00	0.00	3.50	4.03	0.00	\N	\N	\N	{}	2026-07-20 11:43:31.238534
11	5	33	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-21 07:11:31.132643
12	5	26	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-21 07:11:31.132643
13	5	34	\N	1.000	\N	24.9500	0.00	0.00	24.95	28.69	0.00	\N	\N	\N	{}	2026-07-21 07:11:31.132643
14	6	34	\N	1.000	\N	24.9500	0.00	0.00	24.95	28.69	0.00	\N	\N	\N	{}	2026-07-21 07:12:23.915948
15	6	26	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-21 07:12:23.915948
16	6	6	\N	1.000	\N	12.9500	0.00	0.00	12.95	14.89	0.00	\N	\N	\N	{}	2026-07-21 07:12:23.915948
17	7	35	\N	1.000	\N	3.5000	0.00	0.00	3.50	4.03	0.00	\N	\N	\N	{}	2026-07-21 07:12:48.884953
18	7	27	\N	1.000	\N	8.9500	0.00	0.00	8.95	10.29	0.00	\N	\N	\N	{}	2026-07-21 07:12:48.884953
19	7	40	\N	6.000	\N	48.5000	0.00	0.00	291.00	334.65	0.00	\N	\N	\N	{}	2026-07-21 07:12:48.884953
20	8	33	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-22 06:26:08.077662
21	8	35	\N	1.000	\N	3.5000	0.00	0.00	3.50	4.03	0.00	\N	\N	\N	{}	2026-07-22 06:26:08.077662
22	8	6	\N	1.000	\N	12.9500	0.00	0.00	12.95	14.89	0.00	\N	\N	\N	{}	2026-07-22 06:26:08.077662
23	9	34	\N	1.000	\N	24.9500	0.00	0.00	24.95	28.69	0.00	\N	\N	\N	{}	2026-07-22 07:37:06.229289
24	9	6	\N	1.000	\N	12.9500	0.00	0.00	12.95	14.89	0.00	\N	\N	\N	{}	2026-07-22 07:37:06.229289
25	10	40	\N	6.000	\N	48.5000	0.00	0.00	291.00	334.65	0.00	\N	\N	\N	{}	2026-07-22 07:37:39.283604
26	11	21	\N	5.000	\N	85.0000	0.00	0.00	425.00	488.75	0.00	\N	\N	\N	{}	2026-07-22 07:38:22.794235
27	12	22	\N	6.000	\N	18.9500	0.00	0.00	113.70	130.75	0.00	\N	\N	\N	{}	2026-07-22 09:05:40.942959
28	12	22	\N	6.000	\N	18.9500	0.00	0.00	113.70	130.75	0.00	\N	\N	\N	{}	2026-07-22 09:05:40.942959
29	12	22	\N	6.000	\N	18.9500	0.00	0.00	113.70	130.75	0.00	\N	\N	\N	{}	2026-07-22 09:05:40.942959
30	12	3	\N	6.000	\N	14.9500	0.00	0.00	89.70	103.15	0.00	\N	\N	\N	{}	2026-07-22 09:05:40.942959
31	12	22	\N	7.000	\N	18.9500	0.00	0.00	132.65	149.70	0.00	\N	\N	\N	{}	2026-07-22 09:05:40.942959
32	13	22	\N	14.000	\N	18.9500	0.00	0.00	265.30	282.35	0.00	\N	\N	\N	{}	2026-07-22 09:52:19.504137
33	14	22	\N	14.000	\N	18.9500	0.00	0.00	265.30	282.35	0.00	\N	\N	\N	{}	2026-07-22 09:54:51.089491
34	14	35	\N	1.000	\N	3.5000	0.00	0.00	3.50	4.03	0.00	\N	\N	\N	{}	2026-07-22 09:54:51.089491
35	15	33	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-22 10:01:20.577347
36	15	26	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-22 10:01:20.577347
37	15	6	\N	1.000	\N	12.9500	0.00	0.00	12.95	14.89	0.00	\N	\N	\N	{}	2026-07-22 10:01:20.577347
38	16	33	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-22 10:02:26.801934
39	16	26	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-22 10:02:26.801934
40	17	33	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-22 10:02:51.75755
41	18	33	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-22 10:05:41.934214
42	19	33	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-22 10:07:14.54068
43	20	40	\N	1.000	\N	48.5000	0.00	0.00	48.50	55.77	0.00	\N	\N	\N	{}	2026-07-22 13:11:05.680564
44	20	22	\N	6.000	\N	18.9500	0.00	0.00	113.70	130.75	0.00	\N	\N	\N	{}	2026-07-22 13:11:05.680564
45	20	8	\N	10.000	\N	18.0000	0.00	0.00	180.00	207.00	0.00	\N	\N	\N	{}	2026-07-22 13:11:05.680564
46	21	8	\N	10.000	\N	18.0000	0.00	0.00	180.00	207.00	0.00	\N	\N	\N	{}	2026-07-22 13:19:08.05269
47	22	33	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-23 06:18:29.347167
48	22	35	\N	1.000	\N	3.5000	0.00	0.00	3.50	4.03	0.00	\N	\N	\N	{}	2026-07-23 06:18:29.347167
49	22	6	\N	1.000	\N	12.9500	0.00	0.00	12.95	14.89	0.00	\N	\N	\N	{}	2026-07-23 06:18:29.347167
50	22	23	\N	1.000	\N	19.9500	0.00	0.00	19.95	22.94	0.00	\N	\N	\N	{}	2026-07-23 06:18:29.347167
51	23	34	\N	1.000	\N	24.9500	0.00	0.00	24.95	28.69	0.00	\N	\N	\N	{}	2026-07-23 06:28:34.507847
52	23	33	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-23 06:28:34.507847
53	23	35	\N	1.000	\N	3.5000	0.00	0.00	3.50	4.03	0.00	\N	\N	\N	{}	2026-07-23 06:28:34.507847
54	24	33	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-23 06:37:55.643531
55	25	26	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-23 08:31:04.331006
56	25	35	\N	1.000	\N	3.5000	0.00	0.00	3.50	4.03	0.00	\N	\N	\N	{}	2026-07-23 08:31:04.331006
57	25	27	\N	1.000	\N	8.9500	0.00	0.00	8.95	10.29	0.00	\N	\N	\N	{}	2026-07-23 08:31:04.331006
58	26	33	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-23 10:54:04.985485
59	27	22	\N	6.000	\N	18.9500	0.00	0.00	113.70	130.75	0.00	\N	\N	\N	{}	2026-07-23 11:07:35.5777
60	27	22	\N	6.000	\N	18.9500	0.00	0.00	113.70	130.75	0.00	\N	\N	\N	{}	2026-07-23 11:07:35.5777
61	27	40	\N	6.000	\N	48.5000	0.00	0.00	291.00	334.65	0.00	\N	\N	\N	{}	2026-07-23 11:07:35.5777
62	28	34	\N	1.000	\N	24.9500	0.00	0.00	24.95	28.69	0.00	\N	\N	\N	{}	2026-07-23 11:16:48.171078
63	28	26	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-23 11:16:48.171078
64	28	40	\N	1.000	\N	48.5000	0.00	0.00	48.50	55.77	0.00	\N	\N	\N	{}	2026-07-23 11:16:48.171078
65	28	40	\N	3.000	\N	48.5000	0.00	0.00	145.50	167.32	0.00	\N	\N	\N	{}	2026-07-23 11:16:48.171078
66	29	34	\N	1.000	\N	24.9500	0.00	0.00	24.95	28.69	0.00	\N	\N	\N	{}	2026-07-23 11:24:59.699079
67	29	15	\N	1.000	\N	7.9500	0.00	0.00	7.95	9.14	0.00	\N	\N	\N	{}	2026-07-23 11:24:59.699079
68	30	33	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-23 11:42:46.169206
69	31	33	\N	2.000	\N	4.5000	0.00	0.00	9.00	9.67	0.00	\N	\N	\N	{}	2026-07-24 09:14:58.945486
70	31	35	\N	1.000	\N	3.5000	0.00	0.00	3.50	4.03	0.00	\N	\N	\N	{}	2026-07-24 09:14:58.945486
71	31	27	\N	2.000	\N	8.9500	0.00	0.00	17.90	19.24	0.00	\N	\N	\N	{}	2026-07-24 09:14:58.945486
72	31	6	\N	1.000	\N	12.9500	0.00	0.00	12.95	14.89	0.00	\N	\N	\N	{}	2026-07-24 09:14:58.945486
73	31	23	\N	1.000	\N	19.9500	0.00	0.00	19.95	22.94	0.00	\N	\N	\N	{}	2026-07-24 09:14:58.945486
74	32	34	\N	1.000	\N	24.9500	0.00	0.00	24.95	28.69	0.00	\N	\N	\N	{}	2026-07-27 10:47:45.091621
75	33	34	\N	1.000	\N	24.9500	0.00	0.00	24.95	28.69	0.00	\N	\N	\N	{}	2026-07-27 11:06:26.265305
76	33	26	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-27 11:06:26.265305
77	34	34	\N	1.000	\N	24.9500	0.00	0.00	24.95	28.69	0.00	\N	\N	\N	{}	2026-07-27 11:11:19.935793
78	35	17	\N	1.000	\N	14.9500	0.00	0.00	14.95	17.19	0.00	\N	\N	\N	{}	2026-07-27 11:42:20.380769
79	35	29	\N	1.000	\N	3.0000	0.00	0.00	3.00	3.45	0.00	\N	\N	\N	{}	2026-07-27 11:42:20.380769
80	35	33	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-27 11:42:20.380769
81	35	24	\N	1.000	\N	6.5000	0.00	0.00	6.50	7.47	0.00	\N	\N	\N	{}	2026-07-27 11:42:20.380769
82	35	35	\N	8.000	\N	3.5000	0.00	0.00	28.00	32.24	0.00	\N	\N	\N	{}	2026-07-27 11:42:20.380769
83	36	14	\N	6.000	\N	7.9500	0.00	0.00	47.70	54.84	0.00	\N	\N	\N	{}	2026-07-27 11:56:46.179333
84	36	33	\N	43.000	\N	4.5000	0.00	0.00	193.50	222.31	0.00	\N	\N	\N	{}	2026-07-27 11:56:46.179333
85	36	24	\N	1.000	\N	6.5000	0.00	0.00	6.50	7.47	0.00	\N	\N	\N	{}	2026-07-27 11:56:46.179333
86	36	24	\N	1.000	\N	6.5000	0.00	0.00	6.50	7.47	0.00	\N	\N	\N	{}	2026-07-27 11:56:46.179333
87	36	24	\N	1.000	\N	6.5000	0.00	0.00	6.50	7.47	0.00	\N	\N	\N	{}	2026-07-27 11:56:46.179333
88	36	29	\N	1.000	\N	3.0000	0.00	0.00	3.00	3.45	0.00	\N	\N	\N	{}	2026-07-27 11:56:46.179333
89	36	29	\N	1.000	\N	3.0000	0.00	0.00	3.00	3.45	0.00	\N	\N	\N	{}	2026-07-27 11:56:46.179333
90	36	29	\N	1.000	\N	3.0000	0.00	0.00	3.00	3.45	0.00	\N	\N	\N	{}	2026-07-27 11:56:46.179333
91	37	29	\N	8.000	\N	3.0000	0.00	0.00	24.00	27.60	0.00	\N	\N	\N	{}	2026-07-27 12:05:02.757878
92	37	35	\N	39.000	\N	3.5000	0.00	0.00	136.50	157.17	0.00	\N	\N	\N	{}	2026-07-27 12:05:02.757878
93	37	14	\N	55.000	\N	7.9500	0.00	0.00	437.25	502.70	0.00	\N	\N	\N	{}	2026-07-27 12:05:02.757878
94	37	34	\N	2.000	\N	24.9500	0.00	0.00	49.90	57.38	0.00	\N	\N	\N	{}	2026-07-27 12:05:02.757878
95	37	17	\N	3.000	\N	14.9500	0.00	0.00	44.85	51.57	0.00	\N	\N	\N	{}	2026-07-27 12:05:02.757878
96	37	33	\N	7.000	\N	4.5000	0.00	0.00	31.50	36.19	0.00	\N	\N	\N	{}	2026-07-27 12:05:02.757878
97	38	14	\N	5.000	\N	7.9500	0.00	0.00	39.75	45.70	0.00	\N	\N	\N	{}	2026-07-27 12:06:45.411922
98	38	35	\N	8.000	\N	3.5000	0.00	0.00	28.00	32.24	0.00	\N	\N	\N	{}	2026-07-27 12:06:45.411922
99	38	33	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-27 12:06:45.411922
100	38	34	\N	1.000	\N	24.9500	0.00	0.00	24.95	28.69	0.00	\N	\N	\N	{}	2026-07-27 12:06:45.411922
101	38	26	\N	5.000	\N	4.5000	0.00	0.00	22.50	25.85	0.00	\N	\N	\N	{}	2026-07-27 12:06:45.411922
102	39	33	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-28 08:35:54.531925
103	39	27	\N	1.000	\N	8.9500	0.00	0.00	8.95	10.29	0.00	\N	\N	\N	{}	2026-07-28 08:35:54.531925
104	39	35	\N	1.000	\N	3.5000	0.00	0.00	3.50	4.03	0.00	\N	\N	\N	{}	2026-07-28 08:35:54.531925
105	39	6	\N	1.000	\N	12.9500	0.00	0.00	12.95	14.89	0.00	\N	\N	\N	{}	2026-07-28 08:35:54.531925
106	40	33	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-28 08:37:14.679306
107	40	35	\N	1.000	\N	3.5000	0.00	0.00	3.50	4.03	0.00	\N	\N	\N	{}	2026-07-28 08:37:14.679306
108	40	7	\N	1.000	\N	15.5000	0.00	0.00	15.50	17.82	0.00	\N	\N	\N	{}	2026-07-28 08:37:14.679306
109	40	22	\N	1.000	\N	18.9500	0.00	0.00	18.95	21.79	0.00	\N	\N	\N	{}	2026-07-28 08:37:14.679306
110	41	22	\N	6.000	\N	18.9500	0.00	0.00	113.70	130.75	0.00	\N	\N	\N	{}	2026-07-30 08:02:53.794552
111	42	34	\N	1.000	\N	24.9500	0.00	0.00	24.95	28.69	0.00	\N	\N	\N	{}	2026-07-30 08:10:35.386819
112	42	26	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-30 08:10:35.386819
113	43	34	\N	3.000	\N	24.9500	0.00	0.00	74.85	86.07	0.00	\N	\N	\N	{}	2026-07-30 10:59:28.635078
114	43	33	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-30 10:59:28.635078
115	44	33	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-30 11:04:16.977268
116	44	34	\N	1.000	\N	24.9500	0.00	0.00	24.95	28.69	0.00	\N	\N	\N	{}	2026-07-30 11:04:16.977268
117	44	35	\N	1.000	\N	3.5000	0.00	0.00	3.50	4.03	0.00	\N	\N	\N	{}	2026-07-30 11:04:16.977268
118	44	26	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-30 11:04:16.977268
119	44	6	\N	1.000	\N	12.9500	0.00	0.00	12.95	14.89	0.00	\N	\N	\N	{}	2026-07-30 11:04:16.977268
120	44	7	\N	1.000	\N	15.5000	0.00	0.00	15.50	17.82	0.00	\N	\N	\N	{}	2026-07-30 11:04:16.977268
121	44	23	\N	1.000	\N	19.9500	0.00	0.00	19.95	22.94	0.00	\N	\N	\N	{}	2026-07-30 11:04:16.977268
122	44	22	\N	1.000	\N	18.9500	0.00	0.00	18.95	21.79	0.00	\N	\N	\N	{}	2026-07-30 11:04:16.977268
123	44	40	\N	1.000	\N	48.5000	0.00	0.00	48.50	55.77	0.00	\N	\N	\N	{}	2026-07-30 11:04:16.977268
124	44	41	\N	1.000	\N	12.9500	0.00	0.00	12.95	14.89	0.00	\N	\N	\N	{}	2026-07-30 11:04:16.977268
125	44	8	\N	1.000	\N	18.0000	0.00	0.00	18.00	20.70	0.00	\N	\N	\N	{}	2026-07-30 11:04:16.977268
126	44	2	\N	1.000	\N	8.5000	0.00	0.00	8.50	9.78	0.00	\N	\N	\N	{}	2026-07-30 11:04:16.977268
127	44	1	\N	1.000	\N	8.5000	0.00	0.00	8.50	9.78	0.00	\N	\N	\N	{}	2026-07-30 11:04:16.977268
128	44	32	\N	1.000	\N	22.0000	0.00	0.00	22.00	25.30	0.00	\N	\N	\N	{}	2026-07-30 11:04:16.977268
129	44	3	\N	1.000	\N	14.9500	0.00	0.00	14.95	17.19	0.00	\N	\N	\N	{}	2026-07-30 11:04:16.977268
130	44	30	\N	1.000	\N	12.9500	0.00	0.00	12.95	14.89	0.00	\N	\N	\N	{}	2026-07-30 11:04:16.977268
131	44	31	\N	1.000	\N	9.5000	0.00	0.00	9.50	10.93	0.00	\N	\N	\N	{}	2026-07-30 11:04:16.977268
132	44	14	\N	1.000	\N	7.9500	0.00	0.00	7.95	9.14	0.00	\N	\N	\N	{}	2026-07-30 11:04:16.977268
133	44	15	\N	1.000	\N	7.9500	0.00	0.00	7.95	9.14	0.00	\N	\N	\N	{}	2026-07-30 11:04:16.977268
134	44	38	\N	1.000	\N	35.9500	0.00	0.00	35.95	41.34	0.00	\N	\N	\N	{}	2026-07-30 11:04:16.977268
135	44	39	\N	1.000	\N	42.0000	0.00	0.00	42.00	48.30	0.00	\N	\N	\N	{}	2026-07-30 11:04:16.977268
136	44	36	\N	1.000	\N	8.9500	0.00	0.00	8.95	10.29	0.00	\N	\N	\N	{}	2026-07-30 11:04:16.977268
137	44	37	\N	1.000	\N	39.9500	0.00	0.00	39.95	45.94	0.00	\N	\N	\N	{}	2026-07-30 11:04:16.977268
138	44	25	\N	1.000	\N	6.5000	0.00	0.00	6.50	7.47	0.00	\N	\N	\N	{}	2026-07-30 11:04:16.977268
139	44	20	\N	1.000	\N	45.0000	0.00	0.00	45.00	51.75	0.00	\N	\N	\N	{}	2026-07-30 11:04:16.977268
140	44	21	\N	1.000	\N	85.0000	0.00	0.00	85.00	97.75	0.00	\N	\N	\N	{}	2026-07-30 11:04:16.977268
141	44	11	\N	1.000	\N	6.5000	0.00	0.00	6.50	7.47	0.00	\N	\N	\N	{}	2026-07-30 11:04:16.977268
142	44	9	\N	1.000	\N	2.0000	0.00	0.00	2.00	2.30	0.00	\N	\N	\N	{}	2026-07-30 11:04:16.977268
143	45	33	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
144	45	34	\N	1.000	\N	24.9500	0.00	0.00	24.95	28.69	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
145	45	26	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
146	45	35	\N	1.000	\N	3.5000	0.00	0.00	3.50	4.03	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
147	45	27	\N	1.000	\N	8.9500	0.00	0.00	8.95	10.29	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
148	45	6	\N	1.000	\N	12.9500	0.00	0.00	12.95	14.89	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
149	45	23	\N	1.000	\N	19.9500	0.00	0.00	19.95	22.94	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
150	45	7	\N	1.000	\N	15.5000	0.00	0.00	15.50	17.82	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
151	45	22	\N	1.000	\N	18.9500	0.00	0.00	18.95	21.79	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
152	45	40	\N	1.000	\N	48.5000	0.00	0.00	48.50	55.77	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
153	45	8	\N	1.000	\N	18.0000	0.00	0.00	18.00	20.70	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
154	45	41	\N	1.000	\N	12.9500	0.00	0.00	12.95	14.89	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
155	45	1	\N	1.000	\N	8.5000	0.00	0.00	8.50	9.78	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
156	45	2	\N	1.000	\N	8.5000	0.00	0.00	8.50	9.78	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
157	45	32	\N	1.000	\N	22.0000	0.00	0.00	22.00	25.30	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
158	45	3	\N	1.000	\N	14.9500	0.00	0.00	14.95	17.19	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
159	45	30	\N	1.000	\N	12.9500	0.00	0.00	12.95	14.89	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
160	45	31	\N	1.000	\N	9.5000	0.00	0.00	9.50	10.93	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
161	45	14	\N	1.000	\N	7.9500	0.00	0.00	7.95	9.14	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
162	45	15	\N	1.000	\N	7.9500	0.00	0.00	7.95	9.14	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
163	45	39	\N	1.000	\N	42.0000	0.00	0.00	42.00	48.30	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
164	45	38	\N	1.000	\N	35.9500	0.00	0.00	35.95	41.34	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
165	45	36	\N	1.000	\N	8.9500	0.00	0.00	8.95	10.29	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
166	45	25	\N	1.000	\N	6.5000	0.00	0.00	6.50	7.47	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
167	45	24	\N	1.000	\N	6.5000	0.00	0.00	6.50	7.47	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
168	45	20	\N	1.000	\N	45.0000	0.00	0.00	45.00	51.75	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
169	45	11	\N	1.000	\N	6.5000	0.00	0.00	6.50	7.47	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
170	45	29	\N	1.000	\N	3.0000	0.00	0.00	3.00	3.45	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
171	45	10	\N	1.000	\N	2.0000	0.00	0.00	2.00	2.30	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
172	45	17	\N	1.000	\N	14.9500	0.00	0.00	14.95	17.19	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
173	45	19	\N	1.000	\N	32.0000	0.00	0.00	32.00	36.80	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
174	45	16	\N	1.000	\N	12.5000	0.00	0.00	12.50	14.38	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
175	45	13	\N	1.000	\N	1.5000	0.00	0.00	1.50	1.73	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
176	45	5	\N	1.000	\N	3.5000	0.00	0.00	3.50	4.03	0.00	\N	\N	\N	{}	2026-07-30 11:08:54.629521
177	46	5	\N	1.000	\N	3.5000	0.00	0.00	3.50	4.03	0.00	\N	\N	\N	{}	2026-07-31 09:02:44.079433
178	46	5	\N	1.000	\N	3.5000	0.00	0.00	3.50	4.03	0.00	\N	\N	\N	{}	2026-07-31 09:02:44.079433
179	46	8	\N	1.000	\N	18.0000	0.00	0.00	18.00	20.70	0.00	\N	\N	\N	{}	2026-07-31 09:02:44.079433
180	46	5	\N	1.000	\N	3.5000	0.00	0.00	3.50	4.03	0.00	\N	\N	\N	{}	2026-07-31 09:02:44.079433
181	46	7	\N	1.000	\N	15.5000	0.00	0.00	15.50	17.82	0.00	\N	\N	\N	{}	2026-07-31 09:02:44.079433
182	46	5	\N	1.000	\N	3.5000	0.00	0.00	3.50	4.03	0.00	\N	\N	\N	{}	2026-07-31 09:02:44.079433
183	46	5	\N	1.000	\N	3.5000	0.00	0.00	3.50	4.03	0.00	\N	\N	\N	{}	2026-07-31 09:02:44.079433
184	46	5	\N	1.000	\N	3.5000	0.00	0.00	3.50	4.03	0.00	\N	\N	\N	{}	2026-07-31 09:02:44.079433
185	46	5	\N	1.000	\N	3.5000	0.00	0.00	3.50	4.03	0.00	\N	\N	\N	{}	2026-07-31 09:02:44.079433
186	47	5	\N	1.000	\N	3.5000	0.00	0.00	3.50	4.03	0.00	\N	\N	\N	{}	2026-07-31 09:05:05.128956
187	48	33	\N	3.000	\N	4.5000	0.00	0.00	13.50	15.51	0.00	\N	\N	\N	{}	2026-07-31 09:11:02.495422
188	48	34	\N	1.000	\N	24.9500	0.00	0.00	24.95	28.69	0.00	\N	\N	\N	{}	2026-07-31 09:11:02.495422
189	48	35	\N	1.000	\N	3.5000	0.00	0.00	3.50	4.03	0.00	\N	\N	\N	{}	2026-07-31 09:11:02.495422
190	48	26	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-31 09:11:02.495422
191	48	27	\N	1.000	\N	8.9500	0.00	0.00	8.95	10.29	0.00	\N	\N	\N	{}	2026-07-31 09:11:02.495422
192	48	2	\N	1.000	\N	8.5000	0.00	0.00	8.50	9.78	0.00	\N	\N	\N	{}	2026-07-31 09:11:02.495422
193	49	33	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-07-31 09:13:36.251728
194	49	34	\N	1.000	\N	24.9500	0.00	0.00	24.95	28.69	0.00	\N	\N	\N	{}	2026-07-31 09:13:36.251728
195	50	33	\N	2.000	\N	4.5000	0.00	0.00	9.00	10.34	0.00	\N	\N	\N	{}	2026-07-31 10:19:10.764137
196	50	35	\N	1.000	\N	3.5000	0.00	0.00	3.50	4.03	0.00	\N	\N	\N	{}	2026-07-31 10:19:10.764137
197	50	26	\N	2.000	\N	4.5000	0.00	0.00	9.00	10.34	0.00	\N	\N	\N	{}	2026-07-31 10:19:10.764137
198	50	6	\N	2.000	\N	12.9500	0.00	0.00	25.90	29.78	0.00	\N	\N	\N	{}	2026-07-31 10:19:10.764137
199	50	3	\N	1.000	\N	14.9500	0.00	0.00	14.95	17.19	0.00	\N	\N	\N	{}	2026-07-31 10:19:10.764137
200	50	31	\N	1.000	\N	9.5000	0.00	0.00	9.50	10.93	0.00	\N	\N	\N	{}	2026-07-31 10:19:10.764137
201	50	30	\N	2.000	\N	12.9500	0.00	0.00	25.90	29.78	0.00	\N	\N	\N	{}	2026-07-31 10:19:10.764137
202	50	41	\N	1.000	\N	12.9500	0.00	0.00	12.95	14.89	0.00	\N	\N	\N	{}	2026-07-31 10:19:10.764137
203	50	2	\N	1.000	\N	8.5000	0.00	0.00	8.50	9.78	0.00	\N	\N	\N	{}	2026-07-31 10:19:10.764137
204	50	37	\N	2.000	\N	39.9500	0.00	0.00	79.90	91.88	0.00	\N	\N	\N	{}	2026-07-31 10:19:10.764137
205	51	35	\N	1.000	\N	3.5000	0.00	0.00	3.50	4.03	0.00	\N	\N	\N	{}	2026-08-01 06:18:24.476092
206	51	22	\N	1.000	\N	18.9500	0.00	0.00	18.95	21.79	0.00	\N	\N	\N	{}	2026-08-01 06:18:24.476092
207	51	8	\N	1.000	\N	18.0000	0.00	0.00	18.00	20.70	0.00	\N	\N	\N	{}	2026-08-01 06:18:24.476092
208	52	35	\N	4.000	\N	3.5000	0.00	0.00	14.00	16.12	0.00	\N	\N	\N	{}	2026-08-01 06:26:48.17508
209	53	34	\N	1.000	\N	24.9500	0.00	0.00	24.95	28.69	0.00	\N	\N	\N	{}	2026-08-01 06:47:31.609588
210	53	35	\N	1.000	\N	3.5000	0.00	0.00	3.50	4.03	0.00	\N	\N	\N	{}	2026-08-01 06:47:31.609588
211	54	34	\N	10.000	\N	24.9500	0.00	0.00	249.50	286.90	0.00	\N	\N	\N	{}	2026-08-03 05:31:22.847048
212	54	27	\N	1.000	\N	8.9500	0.00	0.00	8.95	10.29	0.00	\N	\N	\N	{}	2026-08-03 05:31:22.847048
213	54	23	\N	5.000	\N	19.9500	0.00	0.00	99.75	114.70	0.00	\N	\N	\N	{}	2026-08-03 05:31:22.847048
214	55	6	\N	1.000	\N	12.9500	0.00	0.00	12.95	14.89	0.00	\N	\N	\N	{}	2026-08-03 09:55:00.686834
215	55	6	\N	1.000	\N	12.9500	0.00	0.00	12.95	14.89	0.00	\N	\N	\N	{}	2026-08-03 09:55:00.686834
216	55	6	\N	1.000	\N	12.9500	0.00	0.00	12.95	14.89	0.00	\N	\N	\N	{}	2026-08-03 09:55:00.686834
217	55	8	\N	1.000	\N	18.0000	0.00	0.00	18.00	20.70	0.00	\N	\N	\N	{}	2026-08-03 09:55:00.686834
218	56	33	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-08-03 10:34:21.541022
219	57	33	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-08-03 10:36:05.378606
220	58	35	\N	1.000	\N	3.5000	0.00	0.00	3.50	4.03	0.00	\N	\N	\N	{}	2026-08-03 10:36:47.672264
221	58	33	\N	1.000	\N	4.5000	0.00	0.00	4.50	5.17	0.00	\N	\N	\N	{}	2026-08-03 10:36:47.672264
222	59	8	\N	1.000	\N	18.0000	0.00	0.00	18.00	20.70	0.00	\N	\N	\N	{}	2026-08-03 11:39:23.355046
223	60	5	\N	1.000	\N	3.5000	0.00	0.00	3.50	4.03	0.00	\N	\N	\N	{}	2026-08-03 11:41:39.528809
224	61	35	\N	2.000	\N	3.5000	0.00	0.00	7.00	8.06	0.00	\N	\N	\N	{}	2026-08-11 11:24:31.15469
225	62	35	\N	1.000	\N	3.5000	0.00	0.00	3.50	4.03	0.00	\N	\N	\N	{}	2026-08-11 11:46:45.440153
226	62	33	\N	2.000	\N	4.5000	0.00	0.00	9.00	10.34	0.00	\N	\N	\N	{}	2026-08-11 11:46:45.440153
227	62	26	\N	3.000	\N	4.5000	0.00	0.00	13.50	15.51	0.00	\N	\N	\N	{}	2026-08-11 11:46:45.440153
228	62	7	\N	1.000	\N	15.5000	0.00	0.00	15.50	17.82	0.00	\N	\N	\N	{}	2026-08-11 11:46:45.440153
\.


--
-- Data for Name: pos_transactions; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.pos_transactions (id, store_id, cashier_id, cashier_session_id, customer_id, pos_terminal_id, transaction_number, transaction_date, transaction_type, subtotal, discount_amount, tax_amount, total_amount, total_cost, amount_paid, change_given, status, price_list_id, sales_order_id, source_cart_id, voided_by, voided_at, metadata, created_at) FROM stdin;
1	1	1	1	1	8	TXN-ORD-20260720082131	2026-07-20 08:21:31.556052	\N	0.00	0.00	0.00	216.78	0.00	0.00	0.00	completed	\N	dd2f97c3-8864-4ea1-92d9-fba5fe3326bf	786d75ea-064f-4ec6-9f1e-15a9cd122adc	\N	\N	{}	2026-07-20 08:21:31.556052
2	1	1	1	1	8	TXN-ORD-20260720083116	2026-07-20 08:31:16.420933	\N	0.00	0.00	0.00	375.41	0.00	0.00	0.00	completed	\N	84b01ab7-92d9-42ba-9741-663c8ecc6b11	1b665464-53d1-4f00-9076-ccc0b9d8d058	\N	\N	{}	2026-07-20 08:31:16.420933
3	1	1	1	1	8	TXN-ORD-20260720110302	2026-07-20 11:03:02.738816	\N	0.00	0.00	0.00	48.75	0.00	0.00	0.00	completed	\N	f31d00e6-3197-4290-8783-5cb985915208	56324b30-f1ce-48c0-aea8-be446dff854c	\N	\N	{}	2026-07-20 11:03:02.738816
4	1	1	4	1	8	TXN-ORD-20260720114331	2026-07-20 11:43:31.236074	\N	0.00	0.00	0.00	32.72	0.00	0.00	0.00	completed	\N	3b0666c1-9ea5-4698-9c8f-06dc39329f39	b6933e63-f839-4d53-a4da-501b113d878a	\N	\N	{}	2026-07-20 11:43:31.236074
5	1	1	6	2	9	TXN-ORD-20260721071131	2026-07-21 07:11:31.128536	\N	0.00	0.00	0.00	39.03	0.00	0.00	0.00	completed	\N	9bc4c20b-4fb5-4c28-9104-8552a3195649	3dfb4094-071e-40a5-bda1-c6e2159e0c7c	\N	\N	{}	2026-07-21 07:11:31.128536
6	1	1	6	1	9	TXN-ORD-20260721071223	2026-07-21 07:12:23.912285	\N	0.00	0.00	0.00	48.75	0.00	0.00	0.00	completed	\N	52a68651-bf78-476f-952a-b29109ba9b2f	ea501529-ce47-4e5f-898a-6c9e2304c995	\N	\N	{}	2026-07-21 07:12:23.912285
7	1	1	6	1	9	TXN-ORD-20260721071248	2026-07-21 07:12:48.881397	\N	0.00	0.00	0.00	348.97	0.00	0.00	0.00	completed	\N	4898af5e-ea60-4761-af8d-08aef13b9274	37447439-4024-451a-8807-b8cfca010487	\N	\N	{}	2026-07-21 07:12:48.881397
8	1	1	8	1	9	TXN-ORD-20260722062608	2026-07-22 06:26:08.073917	\N	0.00	0.00	0.00	24.09	0.00	0.00	0.00	completed	\N	51ce494c-ab3a-4854-bd8e-f5a2cb82c3a5	2a7a15e0-9167-4524-abbd-bfd138bbaa97	\N	\N	{}	2026-07-22 06:26:08.073917
9	1	1	8	2	9	TXN-ORD-20260722073706	2026-07-22 07:37:06.225467	\N	0.00	0.00	0.00	43.58	0.00	0.00	0.00	completed	\N	7d954505-fbde-4d37-8ea3-f733e5f87429	def65158-beb8-4300-83d5-fba82480ba3e	\N	\N	{}	2026-07-22 07:37:06.225467
10	1	1	8	2	9	TXN-ORD-20260722073739	2026-07-22 07:37:39.280079	\N	0.00	0.00	0.00	334.65	0.00	0.00	0.00	completed	\N	80dc64ca-4393-4cf3-897b-2ade3321c8a2	d8d59145-07e8-467c-b016-9555c1af7a88	\N	\N	{}	2026-07-22 07:37:39.280079
11	1	1	8	1	9	TXN-ORD-20260722073822	2026-07-22 07:38:22.791376	\N	0.00	0.00	0.00	488.75	0.00	0.00	0.00	completed	\N	e72fdba4-6d6f-4c2a-8337-606bdca5c876	baccf90f-7f2c-4974-b935-5754c1f8f28a	\N	\N	{}	2026-07-22 07:38:22.791376
12	1	1	10	1	8	TXN-ORD-20260722090540	2026-07-22 09:05:40.939277	\N	0.00	0.00	0.00	645.10	0.00	0.00	0.00	completed	\N	97812e45-0a4d-437b-8641-cf865d76f464	6912df30-1a48-4569-af67-9885dccd1109	\N	\N	{}	2026-07-22 09:05:40.939277
13	1	1	10	10	8	TXN-ORD-20260722095219	2026-07-22 09:52:19.500215	\N	0.00	0.00	0.00	282.35	0.00	0.00	0.00	completed	\N	d3bb4057-5f3d-4eaa-aa7f-2dd14ecfeebb	c03314a4-0319-467c-81c7-32b20fe42417	\N	\N	{}	2026-07-22 09:52:19.500215
14	1	1	10	10	8	TXN-ORD-20260722095451	2026-07-22 09:54:51.086892	\N	0.00	0.00	0.00	286.38	0.00	0.00	0.00	completed	\N	c2fc4fa8-a714-458d-9656-f82085a9c28b	c03314a4-0319-467c-81c7-32b20fe42417	\N	\N	{}	2026-07-22 09:54:51.086892
15	1	1	10	2	8	TXN-ORD-20260722100120	2026-07-22 10:01:20.573258	\N	0.00	0.00	0.00	25.23	0.00	0.00	0.00	completed	\N	dad48529-23c0-4980-8cfc-648f16494e67	161c21c8-61c7-4d8c-b4a0-3c9f7e760e3e	\N	\N	{}	2026-07-22 10:01:20.573258
16	1	1	10	2	8	TXN-ORD-20260722100226	2026-07-22 10:02:26.7996	\N	0.00	0.00	0.00	10.34	0.00	0.00	0.00	completed	\N	98ec425c-3a11-4f1e-ac90-5d96ab6e5819	161c21c8-61c7-4d8c-b4a0-3c9f7e760e3e	\N	\N	{}	2026-07-22 10:02:26.7996
17	1	1	10	10	8	TXN-ORD-20260722100251	2026-07-22 10:02:51.755292	\N	0.00	0.00	0.00	5.17	0.00	0.00	0.00	completed	\N	b5f91a7e-875e-46c5-94d1-c86525cccd5f	6bbea337-0a12-45a9-afb8-0e8eaeeaf9da	\N	\N	{}	2026-07-22 10:02:51.755292
18	1	1	10	10	8	TXN-ORD-20260722100541	2026-07-22 10:05:41.931174	\N	0.00	0.00	0.00	5.17	0.00	0.00	0.00	completed	\N	72138058-5326-4827-a07b-1b92c01446f3	156c1666-36f7-4a09-92e5-22dc40cbeba8	\N	\N	{}	2026-07-22 10:05:41.931174
19	1	1	10	10	8	TXN-ORD-20260722100714	2026-07-22 10:07:14.538155	\N	0.00	0.00	0.00	5.17	0.00	0.00	0.00	completed	\N	489e6dac-d79f-417a-8014-cdc3360080f7	156c1666-36f7-4a09-92e5-22dc40cbeba8	\N	\N	{}	2026-07-22 10:07:14.538155
20	1	1	11	10	8	TXN-ORD-20260722131105	2026-07-22 13:11:05.676768	\N	0.00	0.00	0.00	393.52	0.00	0.00	0.00	completed	\N	11003230-d7b2-44d1-b057-5a5347af254d	f3bc32e2-0d76-4bb6-9ed8-7f02b3021566	\N	\N	{}	2026-07-22 13:11:05.676768
21	1	1	11	10	8	TXN-ORD-20260722131908	2026-07-22 13:19:08.050589	\N	0.00	0.00	0.00	207.00	0.00	0.00	0.00	completed	\N	e04934f5-c4d6-4676-9932-8cfa48b76178	6f23c0ff-e3eb-4ba6-9996-2225ed1afb14	\N	\N	{}	2026-07-22 13:19:08.050589
22	1	1	12	10	9	TXN-ORD-20260723061829	2026-07-23 06:18:29.343689	\N	0.00	0.00	0.00	47.03	0.00	0.00	0.00	completed	\N	bb618d80-6189-4df6-91a2-263755459002	4985babb-0bfa-4532-b61e-63a989fa9ee5	\N	\N	{}	2026-07-23 06:18:29.343689
23	1	1	13	10	9	TXN-ORD-20260723062834	2026-07-23 06:28:34.504164	\N	0.00	0.00	0.00	37.89	0.00	0.00	0.00	completed	\N	fc451cb2-50b8-45d3-80a5-e478c795339a	93837ac6-d0d5-4bb6-b6fe-aae9278f91c1	\N	\N	{}	2026-07-23 06:28:34.504164
24	1	1	15	10	9	TXN-ORD-20260723063755	2026-07-23 06:37:55.640121	\N	0.00	0.00	0.00	5.17	0.00	0.00	0.00	completed	\N	604348ff-bd9a-480f-9d9e-381151944683	e3f11cd0-804c-4c79-95e0-7b910cfacb6d	\N	\N	{}	2026-07-23 06:37:55.640121
25	1	1	16	10	9	TXN-ORD-20260723083104	2026-07-23 08:31:04.327273	\N	0.00	0.00	0.00	19.49	0.00	0.00	0.00	completed	\N	1c15059e-0602-4345-8a9e-3d7aec6962a4	ca973e0c-c7d6-4a65-a74e-895089e0527a	\N	\N	{}	2026-07-23 08:31:04.327273
26	1	1	17	10	9	TXN-ORD-20260723105404	2026-07-23 10:54:04.981812	\N	0.00	0.00	0.00	5.17	0.00	0.00	0.00	completed	\N	275b4bb0-20a8-430e-be00-7e0ee053f49f	0b90118f-6987-47c4-bc66-3325bd219b63	\N	\N	{}	2026-07-23 10:54:04.981812
27	1	1	17	10	8	TXN-ORD-20260723110735	2026-07-23 11:07:35.574065	\N	0.00	0.00	0.00	596.15	0.00	0.00	0.00	completed	\N	25f422ca-343b-4dac-b452-96f9e58dbc82	a0e2078d-2e1b-48a5-896f-ebc458458dc5	\N	\N	{}	2026-07-23 11:07:35.574065
28	1	1	17	1	8	TXN-ORD-20260723111648	2026-07-23 11:16:48.167209	\N	0.00	0.00	0.00	256.95	0.00	0.00	0.00	completed	\N	60f69ca5-9235-44c5-bfd6-acb162e8e2c3	b4ab7b6b-7e3c-43a9-b297-7e4950fd3b96	\N	\N	{}	2026-07-23 11:16:48.167209
29	1	1	17	1	8	TXN-ORD-20260723112459	2026-07-23 11:24:59.696753	\N	0.00	0.00	0.00	37.83	0.00	0.00	0.00	completed	\N	28789b9c-918d-4697-ae9c-19feed267c18	d8c30107-0eb2-4855-bb6a-51cd53505b5d	\N	\N	{}	2026-07-23 11:24:59.696753
30	1	1	18	10	9	TXN-ORD-20260723114246	2026-07-23 11:42:46.164274	\N	0.00	0.00	0.00	5.17	0.00	0.00	0.00	completed	\N	75b6b802-225e-4481-8ef4-cabb0bbfd07c	c269c10a-80e9-41ee-a08b-9c5360c2e4d1	\N	\N	{}	2026-07-23 11:42:46.164274
31	1	1	20	10	9	TXN-ORD-20260724091458	2026-07-24 09:14:58.940278	\N	0.00	0.00	0.00	70.77	0.00	0.00	0.00	completed	\N	2b000f45-99b0-4a8a-b4c7-60dcf39af0b6	bfd00d5b-5d48-491b-ade1-b18de6afcb68	\N	\N	{}	2026-07-24 09:14:58.940278
32	1	1	21	10	8	TXN-ORD-20260727104745	2026-07-27 10:47:45.087737	\N	0.00	0.00	0.00	28.69	0.00	0.00	0.00	completed	\N	17ca1d02-1e68-4055-8370-93d99ef0a469	2eb60586-650e-4927-a018-945a436f01a3	\N	\N	{}	2026-07-27 10:47:45.087737
33	1	1	21	10	9	TXN-ORD-20260727110626	2026-07-27 11:06:26.259804	\N	0.00	0.00	0.00	33.86	0.00	0.00	0.00	completed	\N	2bd4df2f-49f8-409f-aaaa-559039b645fd	c62c4d60-b9e9-4c32-9c09-413b37bed03e	\N	\N	{}	2026-07-27 11:06:26.259804
34	1	1	21	10	9	TXN-ORD-20260727111119	2026-07-27 11:11:19.933547	\N	0.00	0.00	0.00	28.69	0.00	0.00	0.00	completed	\N	d11c1e2b-df97-465b-bff1-34b6780ff4fc	a2eebeb3-662c-46df-82ef-e6ad81058c5b	\N	\N	{}	2026-07-27 11:11:19.933547
35	1	1	22	10	9	TXN-ORD-20260727114220	2026-07-27 11:42:20.376814	\N	0.00	0.00	0.00	65.52	0.00	0.00	0.00	completed	\N	358a5da2-fb5c-496d-b582-a1fd790b4535	14edb196-5851-4397-b37d-66322afe2a2d	\N	\N	{}	2026-07-27 11:42:20.376814
36	1	1	22	10	9	TXN-ORD-20260727115646	2026-07-27 11:56:46.175641	\N	0.00	0.00	0.00	309.91	0.00	0.00	0.00	completed	\N	4aa69799-b182-41e0-9fba-92c284472094	e1dad131-4aa2-487a-b854-e002b43febe4	\N	\N	{}	2026-07-27 11:56:46.175641
37	1	1	22	10	9	TXN-ORD-20260727120502	2026-07-27 12:05:02.755104	\N	0.00	0.00	0.00	832.61	0.00	0.00	0.00	completed	\N	f556b2b5-b3c8-4f70-8ccb-cc09f953a861	334043fe-9a22-4e2d-b316-8b780d6717dd	\N	\N	{}	2026-07-27 12:05:02.755104
38	1	1	22	10	9	TXN-ORD-20260727120645	2026-07-27 12:06:45.405572	\N	0.00	0.00	0.00	137.65	0.00	0.00	0.00	completed	\N	e983e5b0-8b05-4ace-95b8-b629a31013ac	339d0029-bbdf-49c3-8fc1-cadba066f020	\N	\N	{}	2026-07-27 12:06:45.405572
39	1	1	25	10	9	TXN-ORD-20260728083554	2026-07-28 08:35:54.526754	\N	0.00	0.00	0.00	34.38	0.00	0.00	0.00	completed	\N	6fd4ff9a-45d4-4ed7-8002-8b58edc2c9ae	25f07dc1-0c1b-4fc4-a9a5-da5398c129a7	\N	\N	{}	2026-07-28 08:35:54.526754
40	1	1	25	10	9	TXN-ORD-20260728083714	2026-07-28 08:37:14.675685	\N	0.00	0.00	0.00	48.81	0.00	0.00	0.00	completed	\N	2289e488-034e-4715-803e-93febd00c31b	7a73337e-1c3b-4c14-9cd9-9e449851d7ee	\N	\N	{}	2026-07-28 08:37:14.675685
41	10	1	26	10	17	TXN-ORD-20260730080253	2026-07-30 08:02:53.789617	\N	0.00	0.00	0.00	130.75	0.00	0.00	0.00	completed	\N	27b796e5-3c0e-40f5-810e-027abd4a7ae4	62161eda-60af-4495-a576-c6d352e0b040	\N	\N	{}	2026-07-30 08:02:53.789617
42	1	1	27	10	8	TXN-ORD-20260730081035	2026-07-30 08:10:35.384141	\N	0.00	0.00	0.00	33.86	0.00	0.00	0.00	completed	\N	5fd9d2c3-7306-4627-994e-7fbfd52e50a0	252baaaf-5ae9-4d3a-84fe-3e0878ca2b3f	\N	\N	{}	2026-07-30 08:10:35.384141
43	1	1	27	10	8	TXN-ORD-20260730105928	2026-07-30 10:59:28.631111	\N	0.00	0.00	0.00	91.24	0.00	0.00	0.00	completed	\N	5a6f336d-d29c-45b1-955e-e19af7441324	05664a1e-fd9e-4045-81b4-fd5cb0d4110c	\N	\N	{}	2026-07-30 10:59:28.631111
44	1	1	28	10	8	TXN-ORD-20260730110416	2026-07-30 11:04:16.973247	\N	0.00	0.00	0.00	630.62	0.00	0.00	0.00	completed	\N	a4e64219-9b7f-45c4-9a35-106788960f26	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	\N	\N	{}	2026-07-30 11:04:16.973247
45	1	1	28	10	8	TXN-ORD-20260730110854	2026-07-30 11:08:54.625735	\N	0.00	0.00	0.00	582.27	0.00	0.00	0.00	completed	\N	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	aed43720-0c89-4be2-bec4-cbba48662f56	\N	\N	{}	2026-07-30 11:08:54.625735
46	8	1	31	10	12	TXN-ORD-20260731090244	2026-07-31 09:02:44.074907	\N	0.00	0.00	0.00	66.73	0.00	0.00	0.00	completed	\N	d7b077c6-d905-4a48-b75d-96acf66acca6	a6a8a674-e57e-4809-9e86-b7d517ef7d4f	\N	\N	{}	2026-07-31 09:02:44.074907
47	8	1	32	10	12	TXN-ORD-20260731090505	2026-07-31 09:05:05.124705	\N	0.00	0.00	0.00	4.03	0.00	0.00	0.00	completed	\N	bcf0f68b-1976-4186-9409-13d858a952f6	d397fad3-1a15-4db2-a6db-bec5bf2b9297	\N	\N	{}	2026-07-31 09:05:05.124705
48	1	1	31	10	8	TXN-ORD-20260731091102	2026-07-31 09:11:02.490249	\N	0.00	0.00	0.00	73.47	0.00	0.00	0.00	completed	\N	e6c4a620-4b91-4708-bec4-40e04779346b	8f1bca0d-e6d0-434e-9895-8a68343bd13f	\N	\N	{}	2026-07-31 09:11:02.490249
49	1	1	32	10	8	TXN-ORD-20260731091336	2026-07-31 09:13:36.249205	\N	0.00	0.00	0.00	33.86	0.00	0.00	0.00	completed	\N	0fcd3efb-e505-45fe-b224-b8010407fd0e	3469e4f7-1aef-475b-951b-2cd3ec962029	\N	\N	{}	2026-07-31 09:13:36.249205
50	1	6	34	10	8	TXN-ORD-20260731101910	2026-07-31 10:19:10.760242	\N	0.00	0.00	0.00	228.94	0.00	0.00	0.00	completed	\N	8076b3b4-166f-457d-8358-ba2fd13a3832	3b93a016-e219-4b71-a2a8-9ef8886d2aae	\N	\N	{}	2026-07-31 10:19:10.760242
51	1	2	35	10	8	TXN-ORD-20260801061824	2026-08-01 06:18:24.47164	\N	0.00	0.00	0.00	46.52	0.00	0.00	0.00	completed	\N	098814e2-f46e-4be6-9730-87857812e82a	c37191dd-8749-477b-bbe3-402c2e09a31d	\N	\N	{}	2026-08-01 06:18:24.47164
52	1	2	35	10	8	TXN-ORD-20260801062648	2026-08-01 06:26:48.172518	\N	0.00	0.00	0.00	16.12	0.00	0.00	0.00	completed	\N	336b87c3-8860-4889-953b-95145ea1aeab	b627c8a2-bc74-4bb8-9b28-0790aab1266a	\N	\N	{}	2026-08-01 06:26:48.172518
53	1	2	35	10	8	TXN-ORD-20260801064731	2026-08-01 06:47:31.605431	\N	0.00	0.00	0.00	32.72	0.00	0.00	0.00	completed	\N	a7dcdda8-9253-4201-a387-a0276ab91c39	de7f9185-027d-4156-bf2c-7b03cd948e88	\N	\N	{}	2026-08-01 06:47:31.605431
54	1	1	37	10	8	TXN-ORD-20260803053122	2026-08-03 05:31:22.841564	\N	0.00	0.00	0.00	411.89	0.00	0.00	0.00	completed	\N	50388992-d30f-4ee7-929d-ee180ede36a7	7750e699-09d6-45b6-a904-e408f9dabaef	\N	\N	{}	2026-08-03 05:31:22.841564
55	8	1	37	\N	12	TXN-ORD-20260803095500	2026-08-03 09:55:00.682846	\N	0.00	0.00	0.00	65.37	0.00	0.00	0.00	completed	\N	ffc609d0-0015-4ced-9d82-025d4f8a7ca3	b014c45e-d107-4546-9852-0ceafe1d6282	\N	\N	{}	2026-08-03 09:55:00.682846
56	1	1	37	10	8	TXN-ORD-20260803103421	2026-08-03 10:34:21.537172	\N	0.00	0.00	0.00	5.17	0.00	0.00	0.00	completed	\N	8fd7a740-50df-4b24-b468-6eabcd6ec92d	35aa084d-d355-48f6-88d7-8817ee31408b	\N	\N	{}	2026-08-03 10:34:21.537172
57	1	1	37	10	8	TXN-ORD-20260803103605	2026-08-03 10:36:05.376115	\N	0.00	0.00	0.00	5.17	0.00	0.00	0.00	completed	\N	8387e582-d524-4c27-8c35-24ec3f9e8d60	dabc964f-45e9-4234-a23d-4e2fddcdaaee	\N	\N	{}	2026-08-03 10:36:05.376115
58	1	1	37	10	8	TXN-ORD-20260803103647	2026-08-03 10:36:47.66967	\N	0.00	0.00	0.00	9.20	0.00	0.00	0.00	completed	\N	ce52051e-4a6e-440e-909a-86dd0703926d	9f357046-da05-423c-8753-9c3031e83a44	\N	\N	{}	2026-08-03 10:36:47.66967
59	8	1	38	10	12	TXN-ORD-20260803113923	2026-08-03 11:39:23.35089	\N	0.00	0.00	0.00	20.70	0.00	0.00	0.00	completed	\N	480b985a-df97-4c2b-95c3-bef7d77c0e72	73419d52-4182-413f-a4da-14cc77155a65	\N	\N	{}	2026-08-03 11:39:23.35089
60	8	1	38	10	12	TXN-ORD-20260803114139	2026-08-03 11:41:39.526503	\N	0.00	0.00	0.00	4.03	0.00	0.00	0.00	completed	\N	332d8ea2-6299-45fc-b833-ad082299dd72	4620d829-c1f1-4885-9686-3bf3796b3cbb	\N	\N	{}	2026-08-03 11:41:39.526503
61	1	1	39	10	8	TXN-ORD-20260811112431	2026-08-11 11:24:31.146096	\N	0.00	0.00	0.00	8.06	0.00	0.00	0.00	completed	\N	0ad0c821-b76b-4f0c-b149-7edf6df71da9	a0f2df8d-1907-41dd-927d-749d3e09730b	\N	\N	{}	2026-08-11 11:24:31.146096
62	1	1	39	10	8	TXN-ORD-20260811114645	2026-08-11 11:46:45.436115	\N	0.00	0.00	0.00	47.70	0.00	0.00	0.00	completed	\N	1f34855a-1ff6-4e77-a567-65232d4a4504	ecaed928-ee99-437d-8e8d-e46d4597aff8	\N	\N	{}	2026-08-11 11:46:45.436115
\.


--
-- Data for Name: price_lists; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.price_lists (id, name, code, price_list_type, currency_code, valid_from, valid_to, is_default, is_active, metadata, created_at, updated_at) FROM stdin;
1	Retail Price List	RETAIL_SAR	retail	SAR	2024-01-01	\N	t	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
2	Wholesale Price List	WHOLESALE_SAR	wholesale	SAR	2024-01-01	\N	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
3	Promotional Price List	PROMO_SAR	promotional	SAR	2026-07-18	2026-08-17	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:59:03.504911
\.


--
-- Data for Name: product_barcodes; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.product_barcodes (id, product_id, product_variant_id, barcode, barcode_type, is_primary, metadata, created_at) FROM stdin;
1	1	\N	6281000001011	EAN13	t	{}	2026-07-18 07:58:31.573111
2	2	\N	6281000001028	EAN13	t	{}	2026-07-18 07:58:31.573111
3	3	\N	6281030000019	EAN13	t	{}	2026-07-18 07:58:31.573111
4	4	\N	6281000001035	EAN13	t	{}	2026-07-18 07:58:31.573111
5	5	\N	6281007000024	EAN13	t	{}	2026-07-18 07:58:31.573111
6	6	\N	6281000001042	EAN13	t	{}	2026-07-18 07:58:31.573111
7	7	\N	6281000001059	EAN13	t	{}	2026-07-18 07:58:31.573111
8	8	\N	6281000002011	EAN13	t	{}	2026-07-18 07:58:31.573111
9	9	\N	6281055001017	EAN13	t	{}	2026-07-18 07:58:31.573111
10	10	\N	6281055001024	EAN13	t	{}	2026-07-18 07:58:31.573111
11	11	\N	6281055001031	EAN13	t	{}	2026-07-18 07:58:31.573111
12	12	\N	6281000003011	EAN13	t	{}	2026-07-18 07:58:31.573111
13	13	\N	6281000003028	EAN13	t	{}	2026-07-18 07:58:31.573111
14	14	\N	6281000001066	EAN13	t	{}	2026-07-18 07:58:31.573111
15	15	\N	6281000001073	EAN13	t	{}	2026-07-18 07:58:31.573111
16	16	\N	6281000004011	EAN13	t	{}	2026-07-18 07:58:31.573111
17	17	\N	8722700001089	EAN13	t	{}	2026-07-18 07:58:31.573111
18	18	\N	7613035814196	EAN13	t	{}	2026-07-18 07:58:31.573111
19	19	\N	7613035814202	EAN13	t	{}	2026-07-18 07:58:31.573111
20	20	\N	6281000005011	EAN13	t	{}	2026-07-18 07:58:31.573111
21	21	\N	6281000005028	EAN13	t	{}	2026-07-18 07:58:31.573111
22	22	\N	6281000006011	EAN13	t	{}	2026-07-18 07:58:31.573111
23	23	\N	6281000006028	EAN13	t	{}	2026-07-18 07:58:31.573111
24	24	\N	6281000007011	EAN13	t	{}	2026-07-18 07:58:31.573111
25	25	\N	6281000007028	EAN13	t	{}	2026-07-18 07:58:31.573111
26	26	\N	6281000008011	EAN13	t	{}	2026-07-18 07:58:31.573111
27	27	\N	6281000008028	EAN13	t	{}	2026-07-18 07:58:31.573111
28	28	\N	6281000009011	EAN13	t	{}	2026-07-18 07:58:31.573111
29	29	\N	6281000009028	EAN13	t	{}	2026-07-18 07:58:31.573111
30	30	\N	6281006000019	EAN13	t	{}	2026-07-18 07:58:31.573111
31	31	\N	6281006000026	EAN13	t	{}	2026-07-18 07:58:31.573111
32	32	\N	6281000010011	EAN13	t	{}	2026-07-18 07:58:31.573111
33	33	\N	6281001001017	EAN13	t	{}	2026-07-18 07:58:31.573111
34	34	\N	8710908231445	EAN13	t	{}	2026-07-18 07:58:31.573111
35	35	\N	8901030672408	EAN13	t	{}	2026-07-18 07:58:31.573111
36	36	\N	8714789730127	EAN13	t	{}	2026-07-18 07:58:31.573111
37	37	\N	8001841501130	EAN13	t	{}	2026-07-18 07:58:31.573111
38	38	\N	8001841501147	EAN13	t	{}	2026-07-18 07:58:31.573111
39	39	\N	9000101326291	EAN13	t	{}	2026-07-18 07:58:31.573111
40	40	\N	5900627064087	EAN13	t	{}	2026-07-18 07:58:31.573111
41	41	\N	8714789730158	EAN13	t	{}	2026-07-18 07:58:31.573111
42	76	\N	8961008238555	\N	t	\N	2026-07-30 11:30:58.244977
43	77	\N	8886300007268	\N	t	\N	2026-07-30 12:04:13.384973
44	79	1	78797665545454	EAN13	t	{}	2026-08-11 12:21:36.851321
\.


--
-- Data for Name: product_batches; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.product_batches (id, product_id, product_variant_id, batch_number, manufacturing_date, expiry_date, store_id, quantity_available, status, metadata, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: product_categories; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.product_categories (id, parent_category_id, name, code, description, category_level, is_active, metadata, created_at, updated_at) FROM stdin;
1	\N	Food & Groceries	FOOD	Food and grocery items	1	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
2	\N	Beverages	BEVERAGES	All types of beverages	1	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
3	\N	Dairy & Eggs	DAIRY	Dairy products and eggs	1	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
4	\N	Personal Care	PERSONAL_CARE	Personal hygiene and care products	1	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
5	\N	Household & Cleaning	HOUSEHOLD	Household and cleaning supplies	1	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
6	\N	Electronics	ELECTRONICS	Electronic devices and appliances	1	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
7	\N	Bakery	BAKERY	Bread and bakery items	1	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
8	\N	Frozen Foods	FROZEN	Frozen food products	1	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
9	1	Rice & Grains	RICE_GRAINS	Rice, wheat, and grain products	2	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
10	1	Cooking Oil	COOKING_OIL	Cooking and frying oils	2	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
11	1	Canned Foods	CANNED	Canned vegetables, fruits, and ready meals	2	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
12	1	Spices & Seasonings	SPICES	Spices and cooking seasonings	2	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
13	1	Pasta & Noodles	PASTA	Pasta, noodles, and macaroni	2	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
14	1	Sugar & Salt	SUGAR_SALT	Sugar, salt, and sweeteners	2	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
15	2	Soft Drinks	SOFT_DRINKS	Carbonated and non-carbonated soft drinks	2	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
16	2	Juices	JUICES	Fresh and packaged juices	2	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
17	2	Water	WATER	Bottled water and mineral water	2	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
18	2	Tea & Coffee	TEA_COFFEE	Tea, coffee, and hot beverages	2	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
19	2	Energy Drinks	ENERGY_DRINKS	Energy and sports drinks	2	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
20	3	Fresh Milk	FRESH_MILK	Fresh and UHT milk	2	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
21	3	Yogurt & Laban	YOGURT	Yogurt, laban, and cultured dairy	2	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
22	3	Cheese	CHEESE	All types of cheese	2	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
23	3	Butter & Ghee	BUTTER_GHEE	Butter, ghee, and spreads	2	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
24	3	Eggs	EGGS	Fresh eggs	2	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
25	4	Bath & Shower	BATH_SHOWER	Soaps, shower gels, and body wash	2	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
26	4	Hair Care	HAIR_CARE	Shampoo, conditioner, and hair products	2	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
27	4	Oral Care	ORAL_CARE	Toothpaste, toothbrush, and mouthwash	2	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
28	4	Skin Care	SKIN_CARE	Lotions, creams, and skin care	2	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
29	4	Deodorants	DEODORANTS	Deodorants and antiperspirants	2	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
30	5	Laundry Detergents	LAUNDRY	Washing powders and liquid detergents	2	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
31	5	Dishwashing	DISHWASHING	Dishwashing liquid and tablets	2	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
32	5	Surface Cleaners	SURFACE_CLEAN	Floor and surface cleaning products	2	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
33	5	Paper Products	PAPER_PRODUCTS	Tissues, toilet paper, and paper towels	2	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
34	5	Air Fresheners	AIR_FRESH	Air fresheners and deodorizers	2	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
35	8	Frozen Vegetables	FROZEN_VEG	Frozen vegetables and mixed vegetables	2	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
36	8	Frozen Chicken	FROZEN_CHICKEN	Frozen chicken and poultry	2	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
37	8	Frozen Snacks	FROZEN_SNACKS	Frozen snacks and appetizers	2	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
38	8	Ice Cream	ICE_CREAM	Ice cream and frozen desserts	2	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
45	\N	Restaurant Ingredients	REST_ING	Raw materials and ingredients for kitchen	1	t	{}	2026-07-18 08:03:55.324044	2026-07-18 08:03:55.324044
46	45	Proteins	ING_PROTEIN	Meat, poultry, and fish	2	t	{}	2026-07-18 08:04:01.581465	2026-07-18 08:04:01.581465
47	45	Vegetables	ING_VEG	Fresh vegetables	2	t	{}	2026-07-18 08:04:01.581465	2026-07-18 08:04:01.581465
48	45	Dairy & Pantry	ING_DAIRY	Milk, eggs, flour, oil	2	t	{}	2026-07-18 08:04:01.581465	2026-07-18 08:04:01.581465
49	45	Spices & Seasoning	ING_SPICE	Herbs, spices, and sauces	2	t	{}	2026-07-18 08:04:01.581465	2026-07-18 08:04:01.581465
50	\N	Menu Items (Prepared)	MENU_FIN	Finished dishes served to customers	1	t	{}	2026-07-18 08:04:01.581465	2026-07-18 08:04:01.581465
\.


--
-- Data for Name: product_prices; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.product_prices (id, product_id, product_variant_id, price_list_id, uom_id, price, min_quantity, max_quantity, valid_from, valid_to, is_active, metadata, created_at, updated_at) FROM stdin;
1	1	\N	1	3	8.50	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
2	2	\N	1	3	8.50	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
3	3	\N	1	3	14.95	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
4	4	\N	1	3	6.95	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
5	5	\N	1	10	3.50	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
6	6	\N	1	10	12.95	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
7	7	\N	1	10	15.50	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
8	8	\N	1	1	18.00	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
9	9	\N	1	8	2.00	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
10	10	\N	1	8	2.00	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
11	11	\N	1	7	6.50	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
12	12	\N	1	7	1.00	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
13	13	\N	1	7	1.50	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
14	14	\N	1	3	7.95	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
15	15	\N	1	3	7.95	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
16	16	\N	1	4	12.50	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
17	17	\N	1	4	14.95	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
18	18	\N	1	10	28.50	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
19	19	\N	1	10	32.00	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
20	20	\N	1	2	45.00	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
21	21	\N	1	2	85.00	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
22	22	\N	1	3	18.95	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
23	23	\N	1	3	19.95	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
24	24	\N	1	10	6.50	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
25	25	\N	1	10	6.50	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
26	26	\N	1	8	4.50	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
27	27	\N	1	8	8.95	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
28	28	\N	1	2	5.50	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
29	29	\N	1	2	3.00	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
30	30	\N	1	2	12.95	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
31	31	\N	1	10	9.50	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
32	32	\N	1	2	22.00	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
33	33	\N	1	10	4.50	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
34	34	\N	1	11	24.95	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
35	35	\N	1	10	3.50	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
36	36	\N	1	11	8.95	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
37	37	\N	1	2	39.95	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
38	38	\N	1	2	35.95	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
39	39	\N	1	3	42.00	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
40	40	\N	1	4	48.50	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
41	41	\N	1	11	12.95	1.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
42	1	\N	2	3	7.25	12.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
43	3	\N	2	3	12.75	6.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
44	4	\N	2	3	5.95	12.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
45	8	\N	2	1	15.50	10.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
46	9	\N	2	8	1.70	24.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
47	10	\N	2	8	1.70	24.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
48	12	\N	2	7	0.75	24.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
49	16	\N	2	4	10.50	12.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
50	18	\N	2	10	24.50	12.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
51	21	\N	2	2	75.00	5.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
52	22	\N	2	3	16.50	6.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
53	28	\N	2	2	4.75	10.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
54	37	\N	2	2	34.50	6.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
55	38	\N	2	2	30.50	6.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
56	40	\N	2	4	42.00	6.000	\N	\N	\N	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
57	1	\N	3	3	6.99	1.000	\N	2026-07-18	2026-08-17	t	{"promotion_name": "Weekly Special", "discount_percent": 18}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
58	4	\N	3	3	5.50	2.000	\N	2026-07-18	2026-08-17	t	{"promotion_name": "Buy 2 Get Discount", "discount_percent": 21}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
59	9	\N	3	8	1.50	6.000	\N	2026-07-18	2026-08-17	t	{"promotion_name": "6-Pack Deal", "discount_percent": 25}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
60	12	\N	3	7	0.75	12.000	\N	2026-07-18	2026-08-17	t	{"promotion_name": "12-Pack Deal", "discount_percent": 25}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
61	18	\N	3	10	24.99	1.000	\N	2026-07-18	2026-08-17	t	{"promotion_name": "Coffee Week", "discount_percent": 12}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
62	21	\N	3	2	69.99	1.000	\N	2026-07-18	2026-08-17	t	{"promotion_name": "Rice Festival", "discount_percent": 18}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
63	28	\N	3	2	4.99	2.000	\N	2026-07-18	2026-08-17	t	{"promotion_name": "Multi-buy Deal", "discount_percent": 9}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
64	37	\N	3	2	34.99	1.000	\N	2026-07-18	2026-08-17	t	{"promotion_name": "Cleaning Month", "discount_percent": 12}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
65	40	\N	3	4	39.99	1.000	\N	2026-07-18	2026-08-17	t	{"promotion_name": "Cleaning Month", "discount_percent": 18}	2026-07-18 07:59:03.504911	2026-07-18 07:59:03.504911
66	1	\N	2	1	7.50	1.000	\N	\N	\N	t	{"level": "piece", "discount_percent": 11.76}	2026-07-18 08:00:18.458251	2026-07-18 08:00:18.458251
67	1	\N	2	5	85.00	1.000	\N	\N	\N	t	{"level": "carton", "price_per_piece": 7.08, "discount_percent": 16.7}	2026-07-18 08:00:18.458251	2026-07-18 08:00:18.458251
68	9	\N	2	8	1.75	1.000	\N	\N	\N	t	{"level": "can", "discount_percent": 12.5}	2026-07-18 08:00:18.458251	2026-07-18 08:00:18.458251
69	9	\N	2	15	9.90	1.000	\N	\N	\N	t	{"level": "pack", "price_per_can": 1.65, "discount_percent": 17.5}	2026-07-18 08:00:18.458251	2026-07-18 08:00:18.458251
70	9	\N	2	5	36.00	1.000	\N	\N	\N	t	{"level": "carton", "price_per_can": 1.50, "discount_percent": 25}	2026-07-18 08:00:18.458251	2026-07-18 08:00:18.458251
71	12	\N	2	7	0.85	1.000	\N	\N	\N	t	{"level": "bottle"}	2026-07-18 08:00:18.458251	2026-07-18 08:00:18.458251
72	12	\N	2	15	9.00	1.000	\N	\N	\N	t	{"level": "pack", "price_per_bottle": 0.75}	2026-07-18 08:00:18.458251	2026-07-18 08:00:18.458251
73	12	\N	2	5	16.80	1.000	\N	\N	\N	t	{"level": "carton", "price_per_bottle": 0.70}	2026-07-18 08:00:18.458251	2026-07-18 08:00:18.458251
74	8	\N	2	13	16.50	1.000	\N	\N	\N	t	{"level": "tray", "price_per_egg": 0.55}	2026-07-18 08:00:18.458251	2026-07-18 08:00:18.458251
75	8	\N	2	5	180.00	1.000	\N	\N	\N	t	{"level": "carton", "price_per_egg": 0.50, "price_per_tray": 15.00}	2026-07-18 08:00:18.458251	2026-07-18 08:00:18.458251
76	20	\N	2	9	42.00	1.000	\N	\N	\N	t	{"level": "bag", "price_per_kg": 8.40}	2026-07-18 08:00:18.458251	2026-07-18 08:00:18.458251
77	20	\N	2	16	160.00	1.000	\N	\N	\N	t	{"level": "sack", "price_per_kg": 8.00, "price_per_bag": 40.00}	2026-07-18 08:00:18.458251	2026-07-18 08:00:18.458251
78	18	\N	2	1	26.00	1.000	\N	\N	\N	t	{"level": "jar"}	2026-07-18 08:00:18.458251	2026-07-18 08:00:18.458251
79	18	\N	2	5	576.00	1.000	\N	\N	\N	t	{"level": "carton", "price_per_jar": 24.00}	2026-07-18 08:00:18.458251	2026-07-18 08:00:18.458251
83	76	\N	1	6	1.00	1.000	\N	\N	\N	t	\N	2026-07-30 11:30:58.250077	2026-07-30 11:30:58.250077
84	76	\N	1	6	50.00	1.000	0.000	\N	\N	t	{"additionalProp1": {}}	2026-07-30 11:37:44.562099	2026-07-30 11:37:44.562099
85	76	\N	1	4	1000.00	1.000	0.000	\N	\N	t	{"additionalProp1": {}}	2026-07-30 11:37:44.562115	2026-07-30 11:37:44.562115
86	77	\N	1	6	1.00	1.000	\N	\N	\N	t	\N	2026-07-30 12:04:13.389111	2026-07-30 12:04:13.389111
87	77	\N	1	6	50.00	1.000	0.000	\N	\N	t	{"additionalProp1": {}}	2026-07-30 12:08:17.154651	2026-07-30 12:08:17.154651
88	77	\N	1	4	800.00	1.000	0.000	\N	\N	t	{"additionalProp1": {}}	2026-07-30 12:08:17.163828	2026-07-30 12:08:17.163828
89	79	1	1	3	8.00	1.000	0.000	\N	\N	t	{"additionalProp1": {}}	2026-08-11 12:21:36.850408	2026-08-11 12:21:36.850408
90	79	1	1	15	40.00	1.000	0.000	\N	\N	t	{"additionalProp1": {}}	2026-08-11 12:21:36.851067	2026-08-11 12:21:36.851067
91	33	\N	3	10	3.15	1.000	\N	2026-08-10	2026-08-15	t	{"promotion_id": 14, "promotion_name": "NEW PROMO TESTING", "discount_percent": "30%"}	2026-08-11 12:49:59.956495	2026-08-11 12:49:59.956495
93	34	\N	3	11	17.47	1.000	\N	2026-08-10	2026-08-15	t	{"promotion_id": 16, "promotion_name": "Testing SALE01", "discount_percent": "30%"}	2026-08-11 12:49:59.956495	2026-08-11 12:49:59.956495
94	3	\N	3	3	11.96	1.000	\N	2026-08-11	2026-08-15	t	{"promotion_id": 17, "promotion_name": "NEWTEST", "discount_percent": "20%"}	2026-08-11 12:49:59.956495	2026-08-11 12:49:59.956495
\.


--
-- Data for Name: product_serial_numbers; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.product_serial_numbers (id, product_id, product_variant_id, serial_number, status, current_store_id, manufacturing_date, expiry_date, metadata, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: product_uom_conversions; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.product_uom_conversions (id, product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata, created_at) FROM stdin;
1	1	5	1	12.000000	t	{"description": "1 Carton = 12 bottles of 1L milk", "packaging_type": "carton"}	2026-07-18 08:00:18.458251
2	1	1	3	1.000000	t	{"description": "1 Bottle = 1 Liter", "packaging_type": "bottle"}	2026-07-18 08:00:18.458251
3	1	3	11	1000.000000	f	{"description": "1 Liter = 1000 Milliliters"}	2026-07-18 08:00:18.458251
4	2	5	1	12.000000	t	{"description": "1 Carton = 12 bottles", "packaging_type": "carton"}	2026-07-18 08:00:18.458251
5	2	1	3	1.000000	t	{"description": "1 Bottle = 1 Liter", "packaging_type": "bottle"}	2026-07-18 08:00:18.458251
6	3	5	1	6.000000	t	{"description": "1 Carton = 6 bottles of 2L milk", "packaging_type": "carton"}	2026-07-18 08:00:18.458251
7	3	1	3	2.000000	t	{"description": "1 Bottle = 2 Liters", "packaging_type": "bottle"}	2026-07-18 08:00:18.458251
8	4	5	1	12.000000	t	{"description": "1 Carton = 12 bottles", "packaging_type": "carton"}	2026-07-18 08:00:18.458251
9	4	1	3	1.000000	t	{"packaging_type": "bottle"}	2026-07-18 08:00:18.458251
10	5	5	13	4.000000	t	{"description": "1 Carton = 4 trays", "packaging_type": "carton"}	2026-07-18 08:00:18.458251
11	5	13	1	6.000000	t	{"description": "1 Tray = 6 yogurt cups", "packaging_type": "tray"}	2026-07-18 08:00:18.458251
12	5	1	10	170.000000	t	{"description": "1 Cup = 170 grams", "packaging_type": "cup"}	2026-07-18 08:00:18.458251
13	8	13	1	30.000000	t	{"description": "1 Tray = 30 eggs", "packaging_type": "tray"}	2026-07-18 08:00:18.458251
14	8	5	13	12.000000	f	{"description": "1 Carton = 12 trays = 360 eggs", "packaging_type": "carton"}	2026-07-18 08:00:18.458251
15	9	5	15	4.000000	t	{"description": "1 Carton = 4 packs = 24 cans", "packaging_type": "carton"}	2026-07-18 08:00:18.458251
16	9	15	8	6.000000	t	{"description": "1 Pack = 6 cans", "packaging_type": "pack"}	2026-07-18 08:00:18.458251
17	9	8	11	330.000000	t	{"description": "1 Can = 330ml", "packaging_type": "can"}	2026-07-18 08:00:18.458251
18	10	5	15	4.000000	t	{"description": "1 Carton = 4 packs = 24 cans", "packaging_type": "carton"}	2026-07-18 08:00:18.458251
19	10	15	8	6.000000	t	{"description": "1 Pack = 6 cans", "packaging_type": "pack"}	2026-07-18 08:00:18.458251
20	10	8	11	330.000000	t	{"packaging_type": "can"}	2026-07-18 08:00:18.458251
21	11	5	7	8.000000	t	{"description": "1 Carton = 8 bottles of 2L", "packaging_type": "carton"}	2026-07-18 08:00:18.458251
22	11	7	3	2.000000	t	{"description": "1 Bottle = 2 Liters", "packaging_type": "bottle"}	2026-07-18 08:00:18.458251
23	12	5	15	2.000000	t	{"description": "1 Carton = 2 packs = 24 bottles", "packaging_type": "carton"}	2026-07-18 08:00:18.458251
24	12	15	7	12.000000	t	{"description": "1 Pack = 12 bottles", "packaging_type": "pack"}	2026-07-18 08:00:18.458251
25	12	7	11	600.000000	t	{"description": "1 Bottle = 600ml", "packaging_type": "bottle"}	2026-07-18 08:00:18.458251
26	13	5	15	2.000000	t	{"packaging_type": "carton"}	2026-07-18 08:00:18.458251
27	13	15	7	6.000000	t	{"description": "1 Pack = 6 bottles", "packaging_type": "pack"}	2026-07-18 08:00:18.458251
28	13	7	3	1.500000	t	{"description": "1 Bottle = 1.5 Liters", "packaging_type": "bottle"}	2026-07-18 08:00:18.458251
29	14	5	1	12.000000	t	{"description": "1 Carton = 12 bottles", "packaging_type": "carton"}	2026-07-18 08:00:18.458251
30	14	1	3	1.000000	t	{"packaging_type": "bottle"}	2026-07-18 08:00:18.458251
31	15	5	1	12.000000	t	{"packaging_type": "carton"}	2026-07-18 08:00:18.458251
32	15	1	3	1.000000	t	{"packaging_type": "bottle"}	2026-07-18 08:00:18.458251
33	16	5	4	12.000000	t	{"description": "1 Carton = 12 boxes", "packaging_type": "carton"}	2026-07-18 08:00:18.458251
34	16	4	1	100.000000	t	{"description": "1 Box = 100 tea bags", "packaging_type": "box"}	2026-07-18 08:00:18.458251
35	17	5	4	12.000000	t	{"packaging_type": "carton"}	2026-07-18 08:00:18.458251
36	17	4	1	100.000000	t	{"packaging_type": "box"}	2026-07-18 08:00:18.458251
37	18	5	1	24.000000	t	{"description": "1 Carton = 24 jars", "packaging_type": "carton"}	2026-07-18 08:00:18.458251
38	18	1	10	200.000000	t	{"description": "1 Jar = 200g", "packaging_type": "jar"}	2026-07-18 08:00:18.458251
39	19	5	1	24.000000	t	{"packaging_type": "carton"}	2026-07-18 08:00:18.458251
40	19	1	10	200.000000	t	{"packaging_type": "jar"}	2026-07-18 08:00:18.458251
41	20	16	9	4.000000	t	{"description": "1 Sack = 4 bags of 5kg each = 20kg total", "packaging_type": "sack"}	2026-07-18 08:00:18.458251
42	20	9	2	5.000000	t	{"description": "1 Bag = 5kg", "packaging_type": "bag"}	2026-07-18 08:00:18.458251
43	20	2	10	1000.000000	f	{"description": "1 Kilogram = 1000 grams"}	2026-07-18 08:00:18.458251
44	21	16	9	2.000000	t	{"description": "1 Sack = 2 bags of 10kg each", "packaging_type": "sack"}	2026-07-18 08:00:18.458251
45	21	9	2	10.000000	t	{"description": "1 Bag = 10kg", "packaging_type": "bag"}	2026-07-18 08:00:18.458251
46	22	5	7	6.000000	t	{"description": "1 Carton = 6 bottles", "packaging_type": "carton"}	2026-07-18 08:00:18.458251
47	22	7	3	1.800000	t	{"description": "1 Bottle = 1.8L", "packaging_type": "bottle"}	2026-07-18 08:00:18.458251
48	23	5	7	6.000000	t	{"packaging_type": "carton"}	2026-07-18 08:00:18.458251
49	23	7	3	1.800000	t	{"packaging_type": "bottle"}	2026-07-18 08:00:18.458251
50	24	5	6	20.000000	t	{"description": "1 Carton = 20 packets", "packaging_type": "carton"}	2026-07-18 08:00:18.458251
51	24	6	10	500.000000	t	{"description": "1 Packet = 500g", "packaging_type": "packet"}	2026-07-18 08:00:18.458251
52	25	5	6	20.000000	t	{"packaging_type": "carton"}	2026-07-18 08:00:18.458251
53	25	6	10	500.000000	t	{"packaging_type": "packet"}	2026-07-18 08:00:18.458251
54	26	5	8	24.000000	t	{"description": "1 Carton = 24 cans", "packaging_type": "carton"}	2026-07-18 08:00:18.458251
55	26	8	10	400.000000	t	{"description": "1 Can = 400g", "packaging_type": "can"}	2026-07-18 08:00:18.458251
56	27	5	8	48.000000	t	{"description": "1 Carton = 48 small cans", "packaging_type": "carton"}	2026-07-18 08:00:18.458251
57	27	8	10	185.000000	t	{"description": "1 Can = 185g", "packaging_type": "can"}	2026-07-18 08:00:18.458251
58	28	5	9	10.000000	t	{"description": "1 Carton = 10 bags of 1kg", "packaging_type": "carton"}	2026-07-18 08:00:18.458251
59	28	16	2	50.000000	f	{"description": "1 Sack = 50kg bulk sugar", "packaging_type": "sack"}	2026-07-18 08:00:18.458251
60	28	9	2	1.000000	t	{"description": "1 Bag = 1kg", "packaging_type": "bag"}	2026-07-18 08:00:18.458251
61	29	5	9	20.000000	t	{"description": "1 Carton = 20 bags", "packaging_type": "carton"}	2026-07-18 08:00:18.458251
62	29	9	2	1.000000	t	{"packaging_type": "bag"}	2026-07-18 08:00:18.458251
63	30	5	9	10.000000	t	{"description": "1 Carton = 10 bags of frozen fries", "packaging_type": "carton"}	2026-07-18 08:00:18.458251
64	30	9	2	1.000000	t	{"description": "1 Bag = 1kg", "packaging_type": "bag"}	2026-07-18 08:00:18.458251
65	31	5	9	20.000000	t	{"description": "1 Carton = 20 bags", "packaging_type": "carton"}	2026-07-18 08:00:18.458251
66	31	9	10	450.000000	t	{"description": "1 Bag = 450g", "packaging_type": "bag"}	2026-07-18 08:00:18.458251
67	32	5	1	12.000000	t	{"description": "1 Carton = 12 whole chickens", "packaging_type": "carton"}	2026-07-18 08:00:18.458251
68	32	1	2	1.000000	t	{"description": "1 Chicken ≈ 1kg", "packaging_type": "piece"}	2026-07-18 08:00:18.458251
69	33	4	1	48.000000	t	{"description": "1 Box = 48 soap bars", "packaging_type": "box"}	2026-07-18 08:00:18.458251
70	33	1	10	125.000000	t	{"description": "1 Bar = 125g", "packaging_type": "piece"}	2026-07-18 08:00:18.458251
71	34	5	7	12.000000	t	{"description": "1 Carton = 12 bottles", "packaging_type": "carton"}	2026-07-18 08:00:18.458251
72	34	7	11	500.000000	t	{"description": "1 Bottle = 500ml", "packaging_type": "bottle"}	2026-07-18 08:00:18.458251
73	35	4	1	48.000000	t	{"packaging_type": "box"}	2026-07-18 08:00:18.458251
74	35	1	10	120.000000	t	{"packaging_type": "piece"}	2026-07-18 08:00:18.458251
75	36	5	1	24.000000	t	{"description": "1 Carton = 24 tubes", "packaging_type": "carton"}	2026-07-18 08:00:18.458251
76	36	1	11	100.000000	t	{"description": "1 Tube = 100ml", "packaging_type": "tube"}	2026-07-18 08:00:18.458251
77	37	5	9	6.000000	t	{"description": "1 Carton = 6 bags of 3kg", "packaging_type": "carton"}	2026-07-18 08:00:18.458251
78	37	9	2	3.000000	t	{"description": "1 Bag = 3kg", "packaging_type": "bag"}	2026-07-18 08:00:18.458251
79	38	5	9	8.000000	t	{"description": "1 Carton = 8 bags", "packaging_type": "carton"}	2026-07-18 08:00:18.458251
80	38	9	2	2.500000	t	{"description": "1 Bag = 2.5kg", "packaging_type": "bag"}	2026-07-18 08:00:18.458251
81	39	5	7	4.000000	t	{"description": "1 Carton = 4 bottles of 3L", "packaging_type": "carton"}	2026-07-18 08:00:18.458251
82	39	7	3	3.000000	t	{"description": "1 Bottle = 3L", "packaging_type": "bottle"}	2026-07-18 08:00:18.458251
83	40	5	4	6.000000	t	{"description": "1 Carton = 6 boxes", "packaging_type": "carton"}	2026-07-18 08:00:18.458251
84	40	4	1	40.000000	t	{"description": "1 Box = 40 tablets", "packaging_type": "box"}	2026-07-18 08:00:18.458251
85	41	5	7	12.000000	t	{"description": "1 Carton = 12 bottles", "packaging_type": "carton"}	2026-07-18 08:00:18.458251
86	41	7	11	750.000000	t	{"description": "1 Bottle = 750ml", "packaging_type": "bottle"}	2026-07-18 08:00:18.458251
\.


--
-- Data for Name: product_variants; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.product_variants (id, product_id, variant_sku, variant_name, variant_attributes, is_active, metadata, created_at, updated_at) FROM stdin;
1	79	NADEC_FULL_FAT_MILK-250M	NADEC Full Fat Milk - 250ML	{"Size": "250ML"}	t	{}	2026-08-11 12:15:25.966236	2026-08-11 12:15:25.966236
\.


--
-- Data for Name: products; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.products (id, organization_id, sku, name, description, category_id, brand_id, base_uom_id, product_type, tax_category_id, is_serialized, is_batch_managed, is_active, is_sellable, is_purchasable, allow_decimal_quantity, track_inventory, metadata, created_at, updated_at) FROM stdin;
1	1	ALMARAI-MILK-FW-1L	Almarai Fresh Milk Full Fat 1L	Full cream fresh milk	20	1	3	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
2	1	ALMARAI-MILK-LF-1L	Almarai Low Fat Milk 1L	Low fat fresh milk	20	1	3	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
3	1	NADEC-MILK-FW-2L	Nadec Full Cream Milk 2L	Full cream UHT milk 2 liters	20	2	3	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
4	1	ALMARAI-LABAN-1L	Almarai Laban Full Fat 1L	Traditional laban drink	21	1	3	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
5	1	ALSAFI-YOGURT-170G	Al-Safi Greek Yogurt 170g	Greek style yogurt	21	3	10	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
6	1	ALMARAI-CHEESE-SLICE-200G	Almarai Cheese Slices 200g	Processed cheese slices	22	1	10	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
7	1	ALMARAI-FETA-CHEESE-400G	Almarai Feta Cheese 400g	White feta cheese	22	1	10	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
8	1	EGGS-WHITE-30PCS	Fresh White Eggs 30 Pieces	Medium size white eggs tray	24	\N	1	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
9	1	COCA-COLA-330ML	Coca-Cola 330ml Can	Coca-Cola regular can	15	6	8	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
10	1	PEPSI-330ML	Pepsi 330ml Can	Pepsi regular can	15	7	8	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
11	1	COCA-COLA-2L	Coca-Cola 2L Bottle	Coca-Cola 2 liter bottle	15	6	7	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
12	1	WATER-600ML	Bottled Water 600ml	Purified drinking water	17	\N	7	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
13	1	WATER-1.5L	Bottled Water 1.5L	Purified drinking water 1.5 liters	17	\N	7	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
14	1	ALMARAI-ORANGE-1L	Almarai Orange Juice 1L	100% pure orange juice	16	1	3	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
15	1	ALMARAI-MIXED-1L	Almarai Mixed Fruit Juice 1L	Mixed fruit juice	16	1	3	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
16	1	RABEA-TEA-100BAG	Rabea Tea 100 Bags	Premium black tea bags	18	11	4	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
17	1	LIPTON-TEA-100BAG	Lipton Yellow Label Tea 100 Bags	Yellow label tea	18	12	4	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
18	1	NESCAFE-CLASSIC-200G	Nescafe Classic 200g	Instant coffee	18	13	10	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
19	1	NESCAFE-ARABIAN-200G	Nescafe Arabian Coffee 200g	Arabian style instant coffee	18	13	10	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
20	1	RICE-BASMATI-5KG	Basmati Rice 5kg	Premium basmati rice	9	\N	2	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
21	1	RICE-BASMATI-10KG	Basmati Rice 10kg	Premium basmati rice 10kg bag	9	\N	2	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
22	1	OIL-SUNFLOWER-1.8L	Sunflower Cooking Oil 1.8L	Pure sunflower oil	10	\N	3	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
23	1	OIL-CORN-1.8L	Corn Oil 1.8L	Pure corn oil	10	\N	3	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
24	1	PASTA-PENNE-500G	Penne Pasta 500g	Durum wheat penne pasta	13	\N	10	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
25	1	PASTA-SPAGHETTI-500G	Spaghetti Pasta 500g	Durum wheat spaghetti	13	\N	10	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
26	1	CALGARDEN-BEANS-400G	California Garden Baked Beans 400g	Baked beans in tomato sauce	11	10	8	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
27	1	CALGARDEN-TUNA-185G	California Garden Tuna Chunks 185g	Tuna chunks in water	11	10	8	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
28	1	SUGAR-WHITE-1KG	White Sugar 1kg	Refined white sugar	14	\N	2	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
29	1	SALT-TABLE-1KG	Table Salt 1kg	Iodized table salt	14	\N	2	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
30	1	SUNBULAH-FRIES-1KG	Sunbulah French Fries 1kg	Frozen french fries	37	9	2	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
31	1	SUNBULAH-VEGETABLES-450G	Sunbulah Mixed Vegetables 450g	Frozen mixed vegetables	35	9	10	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
32	1	WATANIA-CHICKEN-1KG	Al-Watania Frozen Chicken 1kg	Whole frozen chicken	36	8	2	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
33	1	DETTOL-SOAP-125G	Dettol Original Soap 125g	Antibacterial soap bar	25	15	10	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
34	1	DOVE-BODYWASH-500ML	Dove Body Wash 500ml	Moisturizing body wash	25	19	11	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
35	1	LUX-SOAP-120G	Lux Beauty Soap 120g	Beauty soap bar	25	20	10	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
36	1	PALMOLIVE-TOOTHPASTE-100ML	Palmolive Toothpaste 100ml	Fresh mint toothpaste	27	21	11	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
37	1	TIDE-POWDER-3KG	Tide Washing Powder 3kg	Automatic washing powder	30	16	2	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
38	1	ARIEL-POWDER-2.5KG	Ariel Washing Powder 2.5kg	Automatic washing powder	30	17	2	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
39	1	PERSIL-LIQUID-3L	Persil Liquid Detergent 3L	Liquid laundry detergent	30	18	3	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
40	1	FINISH-TABS-40PCS	Finish Dishwasher Tablets 40pcs	Dishwasher tablets	31	22	4	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
41	1	PALMOLIVE-DISH-750ML	Palmolive Dishwashing Liquid 750ml	Dishwashing liquid	31	21	11	finished_good	1	f	f	t	t	t	f	t	{}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
60	1	ING-CHICKEN-BREAST	Chicken Breast (Fresh)	Fresh chicken breast per kg	46	\N	2	raw_material	1	f	f	t	f	t	f	t	{}	2026-07-18 08:04:09.142891	2026-07-18 08:04:09.142891
61	1	ING-BEEF-MINCED	Minced Beef	Fresh minced beef per kg	46	\N	2	raw_material	1	f	f	t	f	t	f	t	{}	2026-07-18 08:04:09.142891	2026-07-18 08:04:09.142891
62	1	ING-TOMATO	Tomato (Fresh)	Fresh local tomatoes per kg	47	\N	2	raw_material	1	f	f	t	f	t	f	t	{}	2026-07-18 08:04:09.142891	2026-07-18 08:04:09.142891
63	1	ING-ONION	Onion (Red)	Fresh red onions per kg	47	\N	2	raw_material	1	f	f	t	f	t	f	t	{}	2026-07-18 08:04:09.142891	2026-07-18 08:04:09.142891
64	1	ING-LETTUCE	Lettuce	Fresh romaine lettuce per kg	47	\N	2	raw_material	1	f	f	t	f	t	f	t	{}	2026-07-18 08:04:09.142891	2026-07-18 08:04:09.142891
65	1	ING-COOKING-OIL	Cooking Oil (Gallon)	Vegetable cooking oil 5L	48	\N	3	raw_material	1	f	f	t	f	t	f	t	{}	2026-07-18 08:04:09.142891	2026-07-18 08:04:09.142891
66	1	ING-SALT	Cooking Salt	Industrial size cooking salt 10kg	49	\N	2	raw_material	1	f	f	t	f	t	f	t	{}	2026-07-18 08:04:09.142891	2026-07-18 08:04:09.142891
67	1	ING-COFFEE-BEANS	Espresso Beans (Arabica)	Premium Arabica coffee beans per kg	48	\N	2	raw_material	1	f	f	t	f	t	f	t	{}	2026-07-18 08:04:17.41459	2026-07-18 08:04:17.41459
68	1	ING-MILK-REST	Milk (Cafe Grade)	Full cream milk for beverages	48	\N	3	raw_material	1	f	f	t	f	t	f	t	{}	2026-07-18 08:04:17.41459	2026-07-18 08:04:17.41459
76	1	3_IN_1_COFFEE	3 in 1 Coffee	Nescafé 3 in 1 instant coffee mix with coffee, creamer, and sugar. Ready to prepare by adding hot water.	18	13	6	RAW_MATERIAL	1	f	f	t	t	f	t	t	\N	2026-07-30 11:30:58.24077	2026-07-30 11:30:58.24077
77	1	ICED_COFFEE	Iced Coffee	Refreshing iced coffee made with rich coffee flavor, milk, and ice. A chilled beverage perfect for a cool and energizing experience.	18	13	6	RAW_MATERIAL	1	f	f	t	t	f	t	t	\N	2026-07-30 12:04:13.378464	2026-07-30 12:04:13.378464
78	1	FULL_FAT_MILK_1L	Full Fat Milk 1L	NADEC Full Fat Milk 1L is a high-quality UHT full-fat milk, suitable for everyday consumption and rich in essential nutrients.	20	2	3	FINISHED_GOODS	1	f	f	t	t	t	t	t	\N	2026-08-11 12:10:49.320324	2026-08-11 12:10:49.320324
79	1	NADEC_FULL_FAT_MILK	NADEC Full Fat Milk	Full-fat UHT milk by NADEC	20	2	3	FINISHED_GOODS	1	f	f	t	t	t	t	t	\N	2026-08-11 12:14:50.56226	2026-08-11 12:14:50.56226
\.


--
-- Data for Name: profit_loss_analytics; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.profit_loss_analytics (id, organization_id, store_id, date, period_type, month, quarter, year, gross_revenue, sales_discounts, sales_returns, net_revenue, opening_inventory_value, purchases, closing_inventory_value, cogs, gross_profit, gross_profit_margin, total_expenses, net_profit, net_profit_margin, metadata, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: promotions; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.promotions (id, organization_id, code, name, description, promotion_type, action_metadata, valid_from, valid_to, schedule_json, applies_to, target_product_ids, target_category_ids, target_customer_types, min_order_amount, min_quantity, coupon_code, usage_limit, usage_count, usage_per_customer, discount_value, is_stackable, is_active, store_ids, created_by, metadata, created_at, updated_at) FROM stdin;
14	1	PROMO-02	NEW PROMO TESTING	30% ON PRODUCT	percentage_discount	{}	2026-08-10 09:54:00	2026-08-15 09:55:00	null	product	{33}	{}	{}	\N	\N	TEST02	100	0	\N	30.0000	f	t	{1}	1	{}	2026-08-11 09:55:08.855037	2026-08-11 12:49:59.956495
15	1	PROMO34	SALE01	40% ON PRODUCT 34	percentage_discount	{}	2026-08-10 10:00:00	2026-08-20 11:00:00	null	product	{34}	{}	{}	\N	1.000	TEST34	100	0	\N	40.0000	t	t	{1}	1	{}	2026-08-11 10:01:14.438877	2026-08-11 12:49:59.956495
16	1	PROMO-001	Testing SALE01	30% ON PRODUCT	percentage_discount	{}	2026-08-10 10:40:00	2026-08-15 10:40:00	null	product	{34}	{}	{}	\N	1.000	TEST00	100	0	\N	30.0000	f	t	{1}	1	{}	2026-08-11 10:40:23.434314	2026-08-11 12:49:59.956495
17	1	PROMO-3	NEWTEST	20%	percentage_discount	{}	2026-08-11 10:42:00	2026-08-15 10:42:00	null	product	{3}	{}	{}	\N	1.000	TEST3	100	0	\N	20.0000	t	t	{1}	1	{}	2026-08-11 10:42:37.506722	2026-08-11 12:49:59.956495
\.


--
-- Data for Name: purchase_analytics; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.purchase_analytics (id, organization_id, store_id, supplier_id, product_id, category_id, date, month, quarter, year, units_purchased, total_cost, discounts, taxes, net_cost, orders, total_orders, total_quantity, total_amount, discounts_received, taxes_paid, net_amount, average_order_value, metadata, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: purchase_order_lines; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.purchase_order_lines (id, purchase_order_id, product_id, product_variant_id, quantity, uom_id, unit_price, discount_amount, tax_amount, subtotal, line_total, received_quantity, line_number, metadata, created_at) FROM stdin;
\.


--
-- Data for Name: purchase_orders; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.purchase_orders (id, organization_id, po_number, supplier_id, store_id, po_date, expected_delivery_date, status, subtotal, discount_amount, tax_amount, total_amount, price_list_id, created_by, approved_by, metadata, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: quote_lines; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.quote_lines (id, quote_id, organization_id, line_number, product_id, product_variant_id, description, quantity, unit_price, discount_amount, tax_amount, line_total, uom_id, notes, metadata, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: quotes; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.quotes (id, quote_number, organization_id, store_id, customer_id, customer_name, customer_email, customer_phone, quote_status, quote_date, valid_until, sent_date, accepted_date, converted_date, subtotal, discount_amount, tax_amount, total_amount, converted_to_order_id, payment_terms, delivery_terms, terms_and_conditions, notes, internal_notes, created_by_user_id, metadata, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: recipe_ingredients; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.recipe_ingredients (id, recipe_id, product_id, product_variant_id, quantity, uom_id, is_optional, is_byproduct, line_number, metadata, created_at) FROM stdin;
5	3	67	\N	0.018	2	f	f	\N	{}	2026-07-18 08:04:17.41459
6	4	60	\N	0.150	2	f	f	\N	{}	2026-07-18 08:04:17.41459
7	4	62	\N	0.050	2	f	f	\N	{}	2026-07-18 08:04:17.41459
8	4	64	\N	0.030	2	f	f	\N	{}	2026-07-18 08:04:17.41459
\.


--
-- Data for Name: recipes; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.recipes (id, organization_id, recipe_code, recipe_name, description, finished_product_id, yield_quantity, yield_uom_id, preparation_steps, preparation_time_min, cooking_time_min, is_active, metadata, created_at, updated_at) FROM stdin;
3	1	REC-ESPRESSO	Double Espresso	Standard double shot espresso	\N	1.000	1	\N	1	1	t	{}	2026-07-18 08:04:17.41459	2026-07-18 08:04:17.41459
4	1	REC-CLUB-SANDWICH	Classic Club Sandwich	Three layers of chicken, egg, and veg	\N	1.000	1	\N	5	10	t	{}	2026-07-18 08:04:17.41459	2026-07-18 08:04:17.41459
5	1	RG-BF-101	Classic Eggs Benedict	Soft poached eggs, Canadian bacon, and warm hollandaise sauce on toasted English muffins.	76	2.000	1	["Step 1: Melt butter for hollandaise. Whisk egg yolks and lemon juice in a double boiler until thick, then slowly drizzle in the warm butter until emulsified.\\nStep 2: Poach eggs in simmering water with a splash of vinegar for 3-4 minutes until whites are cooked but yolks remain runny.\\nStep 3: Toast the English muffins. Sear the Canadian bacon in a pan for 1 minute per side.\\nStep 4: Assemble by placing bacon on the muffins, topped with poached eggs, and generously drizzle with warm hollandaise sauce."]	15	10	t	[{"qty": 2, "uom": "unit", "variant": "", "unitCost": 1, "productId": 76}, {"qty": 1, "uom": "unit", "variant": "", "unitCost": 1.2, "productId": "13"}]	2026-08-01 11:55:34.454074	2026-08-01 11:55:34.454074
\.


--
-- Data for Name: restaurant_order_items; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.restaurant_order_items (id, order_id, menu_item_id, quantity, unit_price, modifiers_snapshot, modifiers_total, discount_amount, tax_amount, subtotal, line_number, notes, status, metadata, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: restaurant_orders; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.restaurant_orders (id, store_id, table_id, cashier_id, cashier_session_id, customer_id, order_number, order_source, status, subtotal, discount_amount, tax_amount, total_amount, amount_paid, change_given, notes, pos_transaction_id, ordered_at, confirmed_at, served_at, paid_at, metadata, created_at, updated_at) FROM stdin;
2	8	\N	1	38	10	RO-945273	pos	confirmed	10.95	\N	1.64	12.59	\N	\N	\N	\N	2026-08-03 11:19:04.663003	\N	\N	\N	{}	2026-08-03 11:19:04.663774	2026-08-03 11:19:04.663774
3	8	\N	1	38	10	RO-120914	pos	confirmed	45.00	\N	6.75	51.75	\N	\N	\N	\N	2026-08-03 11:22:00.249784	\N	\N	\N	{}	2026-08-03 11:22:00.250606	2026-08-03 11:22:00.250606
4	8	\N	1	38	10	RO-288921	pos	confirmed	12.00	\N	1.80	13.80	\N	\N	\N	\N	2026-08-03 11:41:28.223642	\N	\N	\N	{}	2026-08-03 11:41:28.22457	2026-08-03 11:41:28.22457
\.


--
-- Data for Name: restaurant_tables; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.restaurant_tables (id, store_id, table_number, table_name, section, capacity, is_active, metadata, created_at, updated_at) FROM stdin;
10	8	T-01	Window Table 1	Indoor	2	t	{}	2026-07-18 08:04:17.41459	2026-07-18 08:04:17.41459
11	8	T-02	Window Table 2	Indoor	2	t	{}	2026-07-18 08:04:17.41459	2026-07-18 08:04:17.41459
12	8	T-03	Family Table 1	Indoor	6	t	{}	2026-07-18 08:04:17.41459	2026-07-18 08:04:17.41459
13	8	T-04	Family Table 2	Indoor	6	t	{}	2026-07-18 08:04:17.41459	2026-07-18 08:04:17.41459
14	8	T-05	Booth 1	Indoor	4	t	{}	2026-07-18 08:04:17.41459	2026-07-18 08:04:17.41459
15	8	T-06	Booth 2	Indoor	4	t	{}	2026-07-18 08:04:17.41459	2026-07-18 08:04:17.41459
16	8	T-07	Terrace 1	Outdoor	4	t	{}	2026-07-18 08:04:17.41459	2026-07-18 08:04:17.41459
17	8	T-08	Terrace 2	Outdoor	4	t	{}	2026-07-18 08:04:17.41459	2026-07-18 08:04:17.41459
18	8	T-VIP	VIP Suite 1	VIP	8	t	{}	2026-07-18 08:04:17.41459	2026-07-18 08:04:17.41459
19	8	9	BOFC Table	Terrace	4	t	{}	2026-08-01 12:08:29.402306	2026-08-01 12:08:29.402306
\.


--
-- Data for Name: role_permissions; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.role_permissions (id, role_id, permission_id, scope, metadata, created_at) FROM stdin;
1	1	1	all	{}	2026-07-18 07:59:38.038245
2	1	2	all	{}	2026-07-18 07:59:38.038245
3	1	3	all	{}	2026-07-18 07:59:38.038245
4	1	4	all	{}	2026-07-18 07:59:38.038245
5	1	5	all	{}	2026-07-18 07:59:38.038245
6	1	6	all	{}	2026-07-18 07:59:38.038245
7	1	7	all	{}	2026-07-18 07:59:38.038245
8	1	8	all	{}	2026-07-18 07:59:38.038245
9	1	9	all	{}	2026-07-18 07:59:38.038245
10	1	10	all	{}	2026-07-18 07:59:38.038245
11	1	11	all	{}	2026-07-18 07:59:38.038245
12	1	12	all	{}	2026-07-18 07:59:38.038245
13	1	13	all	{}	2026-07-18 07:59:38.038245
14	1	14	all	{}	2026-07-18 07:59:38.038245
15	1	15	all	{}	2026-07-18 07:59:38.038245
16	1	16	all	{}	2026-07-18 07:59:38.038245
17	1	17	all	{}	2026-07-18 07:59:38.038245
18	1	18	all	{}	2026-07-18 07:59:38.038245
19	1	19	all	{}	2026-07-18 07:59:38.038245
20	1	20	all	{}	2026-07-18 07:59:38.038245
21	1	21	all	{}	2026-07-18 07:59:38.038245
22	1	22	all	{}	2026-07-18 07:59:38.038245
23	1	23	all	{}	2026-07-18 07:59:38.038245
24	1	24	all	{}	2026-07-18 07:59:38.038245
25	1	25	all	{}	2026-07-18 07:59:38.038245
26	1	26	all	{}	2026-07-18 07:59:38.038245
27	1	27	all	{}	2026-07-18 07:59:38.038245
28	1	28	all	{}	2026-07-18 07:59:38.038245
29	1	29	all	{}	2026-07-18 07:59:38.038245
30	1	30	all	{}	2026-07-18 07:59:38.038245
31	1	31	all	{}	2026-07-18 07:59:38.038245
32	1	32	all	{}	2026-07-18 07:59:38.038245
33	1	33	all	{}	2026-07-18 07:59:38.038245
34	1	34	all	{}	2026-07-18 07:59:38.038245
35	1	35	all	{}	2026-07-18 07:59:38.038245
36	1	36	all	{}	2026-07-18 07:59:38.038245
37	1	37	all	{}	2026-07-18 07:59:38.038245
38	1	38	all	{}	2026-07-18 07:59:38.038245
39	1	39	all	{}	2026-07-18 07:59:38.038245
40	1	40	all	{}	2026-07-18 07:59:38.038245
41	1	41	all	{}	2026-07-18 07:59:38.038245
42	1	42	all	{}	2026-07-18 07:59:38.038245
43	1	43	all	{}	2026-07-18 07:59:38.038245
44	1	44	all	{}	2026-07-18 07:59:38.038245
45	1	45	all	{}	2026-07-18 07:59:38.038245
46	1	46	all	{}	2026-07-18 07:59:38.038245
47	1	47	all	{}	2026-07-18 07:59:38.038245
48	1	48	all	{}	2026-07-18 07:59:38.038245
49	1	49	all	{}	2026-07-18 07:59:38.038245
50	1	50	all	{}	2026-07-18 07:59:38.038245
51	1	51	all	{}	2026-07-18 07:59:38.038245
52	1	52	all	{}	2026-07-18 07:59:38.038245
53	1	53	all	{}	2026-07-18 07:59:38.038245
54	1	54	all	{}	2026-07-18 07:59:38.038245
55	1	55	all	{}	2026-07-18 07:59:38.038245
56	1	56	all	{}	2026-07-18 07:59:38.038245
57	1	57	all	{}	2026-07-18 07:59:38.038245
58	1	58	all	{}	2026-07-18 07:59:38.038245
59	1	59	all	{}	2026-07-18 07:59:38.038245
60	1	60	all	{}	2026-07-18 07:59:38.038245
61	1	61	all	{}	2026-07-18 07:59:38.038245
62	1	62	all	{}	2026-07-18 07:59:38.038245
63	1	63	all	{}	2026-07-18 07:59:38.038245
64	1	64	all	{}	2026-07-18 07:59:38.038245
65	1	65	all	{}	2026-07-18 07:59:38.038245
66	1	66	all	{}	2026-07-18 07:59:38.038245
67	1	67	all	{}	2026-07-18 07:59:38.038245
68	1	68	all	{}	2026-07-18 07:59:38.038245
69	1	69	all	{}	2026-07-18 07:59:38.038245
70	1	70	all	{}	2026-07-18 07:59:38.038245
71	1	71	all	{}	2026-07-18 07:59:38.038245
72	1	72	all	{}	2026-07-18 07:59:38.038245
73	1	73	all	{}	2026-07-18 07:59:38.038245
74	2	1	all	{}	2026-07-18 07:59:38.038245
75	2	2	all	{}	2026-07-18 07:59:38.038245
76	2	3	all	{}	2026-07-18 07:59:38.038245
77	2	8	all	{}	2026-07-18 07:59:38.038245
78	2	9	all	{}	2026-07-18 07:59:38.038245
79	2	10	all	{}	2026-07-18 07:59:38.038245
80	2	11	all	{}	2026-07-18 07:59:38.038245
81	2	12	all	{}	2026-07-18 07:59:38.038245
82	2	13	all	{}	2026-07-18 07:59:38.038245
83	2	14	all	{}	2026-07-18 07:59:38.038245
84	2	15	all	{}	2026-07-18 07:59:38.038245
85	2	16	all	{}	2026-07-18 07:59:38.038245
86	2	17	all	{}	2026-07-18 07:59:38.038245
87	2	18	all	{}	2026-07-18 07:59:38.038245
88	2	22	all	{}	2026-07-18 07:59:38.038245
89	2	23	all	{}	2026-07-18 07:59:38.038245
90	2	24	all	{}	2026-07-18 07:59:38.038245
91	2	25	all	{}	2026-07-18 07:59:38.038245
92	2	26	all	{}	2026-07-18 07:59:38.038245
93	2	27	all	{}	2026-07-18 07:59:38.038245
94	2	28	all	{}	2026-07-18 07:59:38.038245
95	2	29	all	{}	2026-07-18 07:59:38.038245
96	2	30	all	{}	2026-07-18 07:59:38.038245
97	2	31	all	{}	2026-07-18 07:59:38.038245
98	2	32	all	{}	2026-07-18 07:59:38.038245
99	2	33	all	{}	2026-07-18 07:59:38.038245
100	2	34	all	{}	2026-07-18 07:59:38.038245
101	2	35	all	{}	2026-07-18 07:59:38.038245
102	2	36	all	{}	2026-07-18 07:59:38.038245
103	2	37	all	{}	2026-07-18 07:59:38.038245
104	2	38	all	{}	2026-07-18 07:59:38.038245
105	2	39	all	{}	2026-07-18 07:59:38.038245
106	2	40	all	{}	2026-07-18 07:59:38.038245
107	2	41	all	{}	2026-07-18 07:59:38.038245
108	2	42	all	{}	2026-07-18 07:59:38.038245
109	2	43	all	{}	2026-07-18 07:59:38.038245
110	2	44	all	{}	2026-07-18 07:59:38.038245
111	2	45	all	{}	2026-07-18 07:59:38.038245
112	2	46	all	{}	2026-07-18 07:59:38.038245
113	2	47	all	{}	2026-07-18 07:59:38.038245
114	2	48	all	{}	2026-07-18 07:59:38.038245
115	2	49	all	{}	2026-07-18 07:59:38.038245
116	2	50	all	{}	2026-07-18 07:59:38.038245
117	2	51	all	{}	2026-07-18 07:59:38.038245
118	2	52	all	{}	2026-07-18 07:59:38.038245
119	2	53	all	{}	2026-07-18 07:59:38.038245
120	2	54	all	{}	2026-07-18 07:59:38.038245
121	2	55	all	{}	2026-07-18 07:59:38.038245
122	2	56	all	{}	2026-07-18 07:59:38.038245
123	2	57	all	{}	2026-07-18 07:59:38.038245
124	2	58	all	{}	2026-07-18 07:59:38.038245
125	2	59	all	{}	2026-07-18 07:59:38.038245
126	2	60	all	{}	2026-07-18 07:59:38.038245
127	2	61	all	{}	2026-07-18 07:59:38.038245
128	2	62	all	{}	2026-07-18 07:59:38.038245
129	2	63	all	{}	2026-07-18 07:59:38.038245
130	2	64	all	{}	2026-07-18 07:59:38.038245
131	2	65	all	{}	2026-07-18 07:59:38.038245
132	2	66	all	{}	2026-07-18 07:59:38.038245
133	2	67	all	{}	2026-07-18 07:59:38.038245
134	2	68	all	{}	2026-07-18 07:59:38.038245
135	2	69	all	{}	2026-07-18 07:59:38.038245
136	2	70	all	{}	2026-07-18 07:59:38.038245
137	2	71	all	{}	2026-07-18 07:59:38.038245
138	2	72	all	{}	2026-07-18 07:59:38.038245
139	2	73	all	{}	2026-07-18 07:59:38.038245
140	3	1	store	{}	2026-07-18 07:59:38.038245
141	3	2	store	{}	2026-07-18 07:59:38.038245
142	3	3	store	{}	2026-07-18 07:59:38.038245
143	3	11	store	{}	2026-07-18 07:59:38.038245
144	3	12	store	{}	2026-07-18 07:59:38.038245
145	3	15	store	{}	2026-07-18 07:59:38.038245
146	3	18	store	{}	2026-07-18 07:59:38.038245
147	3	25	store	{}	2026-07-18 07:59:38.038245
148	3	28	store	{}	2026-07-18 07:59:38.038245
149	3	29	store	{}	2026-07-18 07:59:38.038245
150	3	30	store	{}	2026-07-18 07:59:38.038245
151	3	31	store	{}	2026-07-18 07:59:38.038245
152	3	32	store	{}	2026-07-18 07:59:38.038245
153	3	33	store	{}	2026-07-18 07:59:38.038245
154	3	34	store	{}	2026-07-18 07:59:38.038245
155	3	35	store	{}	2026-07-18 07:59:38.038245
156	3	36	store	{}	2026-07-18 07:59:38.038245
157	3	37	store	{}	2026-07-18 07:59:38.038245
158	3	39	store	{}	2026-07-18 07:59:38.038245
159	3	40	store	{}	2026-07-18 07:59:38.038245
160	3	41	store	{}	2026-07-18 07:59:38.038245
161	3	42	store	{}	2026-07-18 07:59:38.038245
162	3	43	store	{}	2026-07-18 07:59:38.038245
163	3	44	store	{}	2026-07-18 07:59:38.038245
164	3	45	store	{}	2026-07-18 07:59:38.038245
165	3	46	store	{}	2026-07-18 07:59:38.038245
166	3	47	store	{}	2026-07-18 07:59:38.038245
167	3	49	store	{}	2026-07-18 07:59:38.038245
168	3	50	store	{}	2026-07-18 07:59:38.038245
169	3	51	store	{}	2026-07-18 07:59:38.038245
170	3	52	store	{}	2026-07-18 07:59:38.038245
171	3	54	store	{}	2026-07-18 07:59:38.038245
172	3	55	store	{}	2026-07-18 07:59:38.038245
173	3	58	store	{}	2026-07-18 07:59:38.038245
174	3	62	store	{}	2026-07-18 07:59:38.038245
175	3	63	store	{}	2026-07-18 07:59:38.038245
176	3	65	store	{}	2026-07-18 07:59:38.038245
177	3	66	store	{}	2026-07-18 07:59:38.038245
178	3	68	store	{}	2026-07-18 07:59:38.038245
179	3	69	store	{}	2026-07-18 07:59:38.038245
180	3	70	store	{}	2026-07-18 07:59:38.038245
192	5	1	all	{}	2026-07-18 07:59:38.038245
193	5	25	all	{}	2026-07-18 07:59:38.038245
194	5	41	all	{}	2026-07-18 07:59:38.038245
195	5	42	all	{}	2026-07-18 07:59:38.038245
196	5	43	all	{}	2026-07-18 07:59:38.038245
197	5	44	all	{}	2026-07-18 07:59:38.038245
198	5	45	all	{}	2026-07-18 07:59:38.038245
199	5	46	all	{}	2026-07-18 07:59:38.038245
200	5	47	all	{}	2026-07-18 07:59:38.038245
201	5	50	all	{}	2026-07-18 07:59:38.038245
202	5	55	all	{}	2026-07-18 07:59:38.038245
203	5	56	all	{}	2026-07-18 07:59:38.038245
204	5	58	all	{}	2026-07-18 07:59:38.038245
205	5	59	all	{}	2026-07-18 07:59:38.038245
206	5	67	all	{}	2026-07-18 07:59:38.038245
207	5	68	all	{}	2026-07-18 07:59:38.038245
208	5	70	all	{}	2026-07-18 07:59:38.038245
209	6	1	all	{}	2026-07-18 07:59:38.038245
210	6	41	all	{}	2026-07-18 07:59:38.038245
211	6	46	all	{}	2026-07-18 07:59:38.038245
212	6	51	all	{}	2026-07-18 07:59:38.038245
213	6	52	all	{}	2026-07-18 07:59:38.038245
214	6	54	all	{}	2026-07-18 07:59:38.038245
215	6	62	all	{}	2026-07-18 07:59:38.038245
216	6	63	all	{}	2026-07-18 07:59:38.038245
217	6	66	all	{}	2026-07-18 07:59:38.038245
218	6	70	all	{}	2026-07-18 07:59:38.038245
219	7	1	all	{}	2026-07-18 07:59:38.038245
220	7	41	all	{}	2026-07-18 07:59:38.038245
221	7	46	all	{}	2026-07-18 07:59:38.038245
222	7	47	all	{}	2026-07-18 07:59:38.038245
223	7	55	all	{}	2026-07-18 07:59:38.038245
224	7	56	all	{}	2026-07-18 07:59:38.038245
225	7	57	all	{}	2026-07-18 07:59:38.038245
226	7	58	all	{}	2026-07-18 07:59:38.038245
227	7	59	all	{}	2026-07-18 07:59:38.038245
228	7	61	all	{}	2026-07-18 07:59:38.038245
229	7	67	all	{}	2026-07-18 07:59:38.038245
230	7	68	all	{}	2026-07-18 07:59:38.038245
231	7	70	all	{}	2026-07-18 07:59:38.038245
232	8	1	all	{}	2026-07-18 07:59:38.038245
233	8	3	all	{}	2026-07-18 07:59:38.038245
234	8	29	all	{}	2026-07-18 07:59:38.038245
235	8	35	all	{}	2026-07-18 07:59:38.038245
236	8	41	all	{}	2026-07-18 07:59:38.038245
237	8	46	all	{}	2026-07-18 07:59:38.038245
238	8	50	all	{}	2026-07-18 07:59:38.038245
239	8	51	all	{}	2026-07-18 07:59:38.038245
240	8	55	all	{}	2026-07-18 07:59:38.038245
241	8	66	all	{}	2026-07-18 07:59:38.038245
242	8	67	all	{}	2026-07-18 07:59:38.038245
243	8	68	all	{}	2026-07-18 07:59:38.038245
244	8	69	all	{}	2026-07-18 07:59:38.038245
245	8	70	all	{}	2026-07-18 07:59:38.038245
246	1	96	all	"{\\"level\\":\\"admin\\"}"	2026-07-28 09:13:43.882541
247	1	97	all	"{\\"level\\":\\"admin\\"}"	2026-07-29 06:28:03.948084
248	1	103	all	"{\\"level\\":\\"admin\\"}"	2026-07-30 07:54:03.390541
249	1	104	all	"{\\"level\\":\\"admin\\"}"	2026-07-30 07:54:03.39427
250	1	105	All	"{\\"level\\":\\"admin\\"}"	2026-07-30 08:16:00.746725
251	1	106	All	"{\\"level\\":\\"admin\\"}"	2026-07-30 08:16:00.749643
252	1	107	All	"{\\"level\\":\\"admin\\"}"	2026-07-30 08:16:00.75098
253	1	108	All	"{\\"level\\":\\"admin\\"}"	2026-07-30 08:16:00.752439
254	1	109	All	"{\\"level\\":\\"admin\\"}"	2026-07-30 08:16:00.753867
255	1	102	All	"{\\"level\\":\\"admin\\"}"	2026-07-30 10:13:20.243798
256	1	110	All	"{\\"level\\":\\"admin\\"}"	2026-07-31 05:53:55.861135
257	1	111	All	"{\\"level\\":\\"admin\\"}"	2026-07-31 05:53:55.864446
258	1	112	All	"{\\"level\\":\\"admin\\"}"	2026-07-31 05:53:55.866037
259	1	113	All	"{\\"level\\":\\"admin\\"}"	2026-07-31 05:53:55.867612
260	1	114	All	"{\\"level\\":\\"admin\\"}"	2026-07-31 05:53:55.869132
261	1	101	All	"{\\"level\\":\\"admin\\"}"	2026-07-31 05:55:20.29153
262	1	100	All	"{\\"level\\":\\"admin\\"}"	2026-07-31 05:55:20.294355
267	1	115	All	"{\\"level\\":\\"admin\\"}"	2026-07-31 12:14:50.861488
268	1	116	All	"{\\"level\\":\\"admin\\"}"	2026-07-31 12:14:50.863715
269	1	117	All	"{\\"level\\":\\"admin\\"}"	2026-07-31 12:14:50.865275
270	1	118	All	"{\\"level\\":\\"admin\\"}"	2026-07-31 12:14:50.866785
271	1	119	All	"{\\"level\\":\\"admin\\"}"	2026-07-31 12:14:50.868313
272	1	120	All	"{\\"level\\":\\"admin\\"}"	2026-07-31 12:14:50.869783
273	1	90	All	"{\\"level\\":\\"admin\\"}"	2026-07-31 12:20:27.200758
274	1	89	All	"{\\"level\\":\\"admin\\"}"	2026-07-31 12:20:27.20257
279	1	91	All	"{\\"level\\":\\"admin\\"}"	2026-08-01 08:02:06.088064
280	1	92	All	"{\\"level\\":\\"admin\\"}"	2026-08-01 08:04:36.772858
288	1	121	All	"{\\"level\\":\\"admin\\"}"	2026-08-11 06:45:10.207531
290	4	122	all	"{\\"level\\":\\"admin\\"}"	2026-08-11 07:15:56.085182
291	4	123	all	"{\\"level\\":\\"admin\\"}"	2026-08-11 07:17:21.220576
292	4	124	all	"{\\"level\\":\\"admin\\"}"	2026-08-11 07:17:21.224423
293	4	125	all	"{\\"level\\":\\"admin\\"}"	2026-08-11 07:17:21.226254
294	4	126	all	"{\\"level\\":\\"admin\\"}"	2026-08-11 07:21:10.970102
295	4	127	all	"{\\"level\\":\\"admin\\"}"	2026-08-11 07:21:10.971865
296	4	128	all	"{\\"level\\":\\"admin\\"}"	2026-08-11 07:22:58.225095
297	4	129	all	"{\\"level\\":\\"admin\\"}"	2026-08-11 07:23:24.460301
298	4	130	all	"{\\"level\\":\\"admin\\"}"	2026-08-11 07:23:58.527023
299	4	131	all	"{\\"level\\":\\"admin\\"}"	2026-08-11 07:25:40.376841
300	4	132	all	"{\\"level\\":\\"admin\\"}"	2026-08-11 07:25:40.379185
301	4	133	all	"{\\"level\\":\\"admin\\"}"	2026-08-11 07:25:40.381032
302	4	134	all	"{\\"level\\":\\"admin\\"}"	2026-08-11 07:34:38.263623
303	4	135	all	"{\\"level\\":\\"admin\\"}"	2026-08-11 07:34:38.265631
304	4	136	all	"{\\"level\\":\\"admin\\"}"	2026-08-11 07:34:38.267641
305	4	137	all	"{\\"level\\":\\"admin\\"}"	2026-08-11 07:34:38.269624
306	4	138	all	"{\\"level\\":\\"admin\\"}"	2026-08-11 07:34:38.271575
307	4	139	all	"{\\"level\\":\\"admin\\"}"	2026-08-11 07:34:38.273544
308	4	141	all	"{\\"level\\":\\"admin\\"}"	2026-08-11 07:36:04.472522
309	4	142	all	"{\\"level\\":\\"admin\\"}"	2026-08-11 07:36:04.474117
310	4	143	all	"{\\"level\\":\\"admin\\"}"	2026-08-11 07:37:23.490254
311	4	144	all	"{\\"level\\":\\"admin\\"}"	2026-08-11 07:37:23.493438
312	4	145	all	"{\\"level\\":\\"admin\\"}"	2026-08-11 07:48:09.254064
313	4	146	all	"{\\"level\\":\\"admin\\"}"	2026-08-11 07:48:49.013849
314	4	147	all	"{\\"level\\":\\"admin\\"}"	2026-08-11 07:48:49.015503
315	4	148	all	"{\\"level\\":\\"admin\\"}"	2026-08-11 07:49:24.150704
316	4	149	all	"{\\"level\\":\\"admin\\"}"	2026-08-11 07:50:05.702842
317	4	150	all	"{\\"level\\":\\"admin\\"}"	2026-08-11 07:50:47.705395
318	4	151	all	"{\\"level\\":\\"admin\\"}"	2026-08-11 07:53:14.019873
319	9	33	pos:apply_discounts	{"notes": "permission 1"}	2026-08-11 07:59:51.129243
320	9	34	pos:process_returns	{"notes": "permission 2"}	2026-08-11 07:59:51.131003
321	9	102	promotions_&_discounts:list	{"notes": "permission 3"}	2026-08-11 07:59:51.132284
322	9	101	promotions_&_discounts:add	{"notes": "permission 4"}	2026-08-11 07:59:51.13359
323	9	92	restaurant:process_orders	{"notes": "permission 5"}	2026-08-11 07:59:51.134841
\.


--
-- Data for Name: role_ui_customizations; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.role_ui_customizations (id, role_id, submenu_id, customization_data, metadata, created_at, updated_at) FROM stdin;
1	1	1	{"layout": "grid", "widgets": ["sales_overview", "inventory_status", "recent_transactions", "low_stock_alerts", "top_products"], "refresh_interval": 30}	{}	2026-07-18 07:59:38.038245	2026-07-18 07:59:38.038245
2	2	1	{"layout": "grid", "widgets": ["sales_overview", "profit_analysis", "store_performance", "inventory_value"], "refresh_interval": 60}	{}	2026-07-18 07:59:38.038245	2026-07-18 07:59:38.038245
3	3	2	{"layout": "list", "widgets": ["daily_sales", "active_cashiers", "inventory_alerts", "pending_orders"], "refresh_interval": 30}	{}	2026-07-18 07:59:38.038245	2026-07-18 07:59:38.038245
4	4	22	{"barcode_scanner": true, "max_discount_percent": 10, "quick_access_categories": true, "discount_requires_approval": true}	{}	2026-07-18 07:59:38.038245	2026-07-18 07:59:38.038245
\.


--
-- Data for Name: roles; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.roles (id, name, code, description, is_system_role, is_active, metadata, created_at, updated_at) FROM stdin;
1	Super Administrator	super_admin	Full system access with all permissions including tenant management	t	t	{"scope": "all"}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
2	Owner	owner	Organization owner with full access except tenant and UI module management	t	t	{"scope": "all"}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
3	Store Manager	store_manager	Manages store operations, inventory, sales, and staff	f	t	{"scope": "all"}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
4	Cashier	cashier	Processes sales transactions at POS	f	t	{"scope": "own"}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
5	Inventory Manager	inventory_manager	Manages inventory, stock counts, and transfers	f	t	{"scope": "own"}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
6	Sales Executive	sales_executive	Manages customers and sales orders	f	t	{"scope": "own"}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
7	Purchase Manager	purchase_manager	Manages suppliers and purchase orders	f	t	{"scope": "own"}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
8	Accountant	accountant	Access to financial reports and analytics	f	t	{"scope": "own"}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
9	Sale Cashier	sale_cashier	for new module	f	t	{"scope": "all"}	2026-08-11 07:59:49.818845	2026-08-11 07:59:49.818845
\.


--
-- Data for Name: sales_analytics; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.sales_analytics (id, organization_id, store_id, product_id, category_id, customer_id, date, hour, day_of_week, month, quarter, year, units_sold, revenue, discounts, taxes, net_revenue, transactions, payment_method, payment_gateway, average_order_value, metadata, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: sales_order_lines; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.sales_order_lines (id, sales_order_id, product_id, product_variant_id, quantity, uom_id, unit_price, discount_amount, tax_amount, subtotal, line_total, shipped_quantity, line_number, metadata, created_at) FROM stdin;
\.


--
-- Data for Name: sales_order_lines_v2; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.sales_order_lines_v2 (id, sales_order_id, organization_id, line_number, product_id, product_variant_id, product_name, product_sku, quantity_ordered, quantity_fulfilled, quantity_cancelled, quantity_returned, uom_id, unit_price, discount_amount, discount_percentage, tax_amount, line_total, tax_category_id, tax_rate, batch_number, serial_numbers, expiry_date, line_status, customization_details, unit_cost, notes, metadata, created_at, updated_at) FROM stdin;
5305e57f-4336-4fc7-ade3-40f8013942a1	dd2f97c3-8864-4ea1-92d9-fba5fe3326bf	1	1	8	\N	Fresh White Eggs 30 Pieces	EGGS-WHITE-30PCS	10.000	0.000	0.000	0.000	1	18.00	\N	\N	27.00	207.00	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-20 08:21:31.529725	2026-07-20 08:21:31.529725
078637d7-9c70-48e9-8498-e1e1c4a28fb2	dd2f97c3-8864-4ea1-92d9-fba5fe3326bf	1	2	2	\N	Almarai Low Fat Milk 1L	ALMARAI-MILK-LF-1L	1.000	0.000	0.000	0.000	3	8.50	\N	\N	1.27	9.78	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-20 08:21:31.552289	2026-07-20 08:21:31.552289
2ab5c658-f421-4c98-a9aa-bbd9da6d0517	84b01ab7-92d9-42ba-9741-663c8ecc6b11	1	1	40	\N	Finish Dishwasher Tablets 40pcs	FINISH-TABS-40PCS	6.000	0.000	0.000	0.000	4	48.50	\N	\N	43.65	334.65	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-20 08:31:16.405848	2026-07-20 08:31:16.405848
e581342e-ae81-4604-95cc-7b09f9a8d6d7	84b01ab7-92d9-42ba-9741-663c8ecc6b11	1	2	23	\N	Corn Oil 1.8L	OIL-CORN-1.8L	1.000	0.000	0.000	0.000	3	19.95	\N	\N	2.99	22.94	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-20 08:31:16.414537	2026-07-20 08:31:16.414537
61c15c47-971e-4fba-90fa-1c11b63b8e7b	84b01ab7-92d9-42ba-9741-663c8ecc6b11	1	3	7	\N	Almarai Feta Cheese 400g	ALMARAI-FETA-CHEESE-400G	1.000	0.000	0.000	0.000	10	15.50	\N	\N	2.32	17.82	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-20 08:31:16.417673	2026-07-20 08:31:16.417673
a31e1c0a-fb4f-4152-b6b9-03ae6933dadd	f31d00e6-3197-4290-8783-5cb985915208	1	1	34	\N	Dove Body Wash 500ml	DOVE-BODYWASH-500ML	1.000	0.000	0.000	0.000	11	24.95	\N	\N	3.74	28.69	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-20 11:03:02.722886	2026-07-20 11:03:02.722886
f950c93a-3666-45de-8e22-b772c447276c	f31d00e6-3197-4290-8783-5cb985915208	1	2	26	\N	California Garden Baked Beans 400g	CALGARDEN-BEANS-400G	1.000	0.000	0.000	0.000	8	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-20 11:03:02.731844	2026-07-20 11:03:02.731844
8621840c-a049-4c8a-8cbc-6e1d445fa987	f31d00e6-3197-4290-8783-5cb985915208	1	3	6	\N	Almarai Cheese Slices 200g	ALMARAI-CHEESE-SLICE-200G	1.000	0.000	0.000	0.000	10	12.95	\N	\N	1.94	14.89	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-20 11:03:02.735033	2026-07-20 11:03:02.735033
128759b0-21b8-47c6-8d19-b5723e2013c7	3b0666c1-9ea5-4698-9c8f-06dc39329f39	1	1	34	\N	Dove Body Wash 500ml	DOVE-BODYWASH-500ML	1.000	0.000	0.000	0.000	11	24.95	\N	\N	3.74	28.69	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-20 11:43:31.229598	2026-07-20 11:43:31.229598
3f7c9833-ad5a-4f2f-bf28-77ec34735804	3b0666c1-9ea5-4698-9c8f-06dc39329f39	1	2	35	\N	Lux Beauty Soap 120g	LUX-SOAP-120G	1.000	0.000	0.000	0.000	10	3.50	\N	\N	0.53	4.03	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-20 11:43:31.233122	2026-07-20 11:43:31.233122
a35a23bd-cc19-4358-8285-ed01321bea9f	9bc4c20b-4fb5-4c28-9104-8552a3195649	1	1	33	\N	Dettol Original Soap 125g	DETTOL-SOAP-125G	1.000	0.000	0.000	0.000	10	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-21 07:11:31.112697	2026-07-21 07:11:31.112697
c5a56c3a-6717-4d1b-8963-8785c9bf1ccc	9bc4c20b-4fb5-4c28-9104-8552a3195649	1	2	26	\N	California Garden Baked Beans 400g	CALGARDEN-BEANS-400G	1.000	0.000	0.000	0.000	8	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-21 07:11:31.121976	2026-07-21 07:11:31.121976
48ce660b-fffd-4994-ae5a-36849b74b7db	9bc4c20b-4fb5-4c28-9104-8552a3195649	1	3	34	\N	Dove Body Wash 500ml	DOVE-BODYWASH-500ML	1.000	0.000	0.000	0.000	11	24.95	\N	\N	3.74	28.69	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-21 07:11:31.12512	2026-07-21 07:11:31.12512
ec61e58c-227f-4483-a7a6-f1a0f992fb02	52a68651-bf78-476f-952a-b29109ba9b2f	1	1	34	\N	Dove Body Wash 500ml	DOVE-BODYWASH-500ML	1.000	0.000	0.000	0.000	11	24.95	\N	\N	3.74	28.69	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-21 07:12:23.897262	2026-07-21 07:12:23.897262
effc33fd-a42f-4b25-bd55-b1c5d6637759	52a68651-bf78-476f-952a-b29109ba9b2f	1	2	26	\N	California Garden Baked Beans 400g	CALGARDEN-BEANS-400G	1.000	0.000	0.000	0.000	8	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-21 07:12:23.905909	2026-07-21 07:12:23.905909
f2c98948-9f3c-40ed-953d-299fab34cc27	52a68651-bf78-476f-952a-b29109ba9b2f	1	3	6	\N	Almarai Cheese Slices 200g	ALMARAI-CHEESE-SLICE-200G	1.000	0.000	0.000	0.000	10	12.95	\N	\N	1.94	14.89	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-21 07:12:23.909058	2026-07-21 07:12:23.909058
6eba69ac-8f5f-4112-b13e-69ffca7760a7	4898af5e-ea60-4761-af8d-08aef13b9274	1	1	35	\N	Lux Beauty Soap 120g	LUX-SOAP-120G	1.000	0.000	0.000	0.000	10	3.50	\N	\N	0.53	4.03	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-21 07:12:48.867624	2026-07-21 07:12:48.867624
7f5623ef-9940-4873-b398-9d497b18ff33	4898af5e-ea60-4761-af8d-08aef13b9274	1	2	27	\N	California Garden Tuna Chunks 185g	CALGARDEN-TUNA-185G	1.000	0.000	0.000	0.000	8	8.95	\N	\N	1.34	10.29	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-21 07:12:48.875032	2026-07-21 07:12:48.875032
26913bca-b300-4029-ab98-89ae230e2c5b	4898af5e-ea60-4761-af8d-08aef13b9274	1	3	40	\N	Finish Dishwasher Tablets 40pcs	FINISH-TABS-40PCS	6.000	0.000	0.000	0.000	4	48.50	\N	\N	43.65	334.65	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-21 07:12:48.878174	2026-07-21 07:12:48.878174
b7160c56-7aed-4a14-9ba1-d40bc1af1f76	51ce494c-ab3a-4854-bd8e-f5a2cb82c3a5	1	1	33	\N	Dettol Original Soap 125g	DETTOL-SOAP-125G	1.000	0.000	0.000	0.000	10	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-22 06:26:08.060627	2026-07-22 06:26:08.060627
c74d65a2-4e32-4a03-bc00-48080d847eae	51ce494c-ab3a-4854-bd8e-f5a2cb82c3a5	1	2	35	\N	Lux Beauty Soap 120g	LUX-SOAP-120G	1.000	0.000	0.000	0.000	10	3.50	\N	\N	0.53	4.03	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-22 06:26:08.067873	2026-07-22 06:26:08.067873
3d6935c4-d1f5-4e46-8cb9-474e363b3133	51ce494c-ab3a-4854-bd8e-f5a2cb82c3a5	1	3	6	\N	Almarai Cheese Slices 200g	ALMARAI-CHEESE-SLICE-200G	1.000	0.000	0.000	0.000	10	12.95	\N	\N	1.94	14.89	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-22 06:26:08.070729	2026-07-22 06:26:08.070729
c71ba242-75aa-4836-a144-ed7e00704697	7d954505-fbde-4d37-8ea3-f733e5f87429	1	1	34	\N	Dove Body Wash 500ml	DOVE-BODYWASH-500ML	1.000	0.000	0.000	0.000	11	24.95	\N	\N	3.74	28.69	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-22 07:37:06.213954	2026-07-22 07:37:06.213954
0ea47ebb-12fb-4b38-ae13-9428b227a150	7d954505-fbde-4d37-8ea3-f733e5f87429	1	2	6	\N	Almarai Cheese Slices 200g	ALMARAI-CHEESE-SLICE-200G	1.000	0.000	0.000	0.000	10	12.95	\N	\N	1.94	14.89	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-22 07:37:06.222227	2026-07-22 07:37:06.222227
5eac0b7f-e0f8-4793-bdf7-92dfa3612cce	80dc64ca-4393-4cf3-897b-2ade3321c8a2	1	1	40	\N	Finish Dishwasher Tablets 40pcs	FINISH-TABS-40PCS	6.000	0.000	0.000	0.000	4	48.50	\N	\N	43.65	334.65	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-22 07:37:39.271677	2026-07-22 07:37:39.271677
6b76b796-abf4-402a-b104-f5b242bfe4cb	e72fdba4-6d6f-4c2a-8337-606bdca5c876	1	1	21	\N	Basmati Rice 10kg	RICE-BASMATI-10KG	5.000	0.000	0.000	0.000	2	85.00	\N	\N	63.75	488.75	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-22 07:38:22.787444	2026-07-22 07:38:22.787444
6b49bc51-3281-43f6-b01f-c2d34619b8d6	97812e45-0a4d-437b-8641-cf865d76f464	1	1	22	\N	Sunflower Cooking Oil 1.8L	OIL-SUNFLOWER-1.8L	6.000	0.000	0.000	0.000	3	18.95	\N	\N	17.05	130.75	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-22 09:05:40.916678	2026-07-22 09:05:40.916678
40db0066-fce5-49a7-b9c6-caf037c50722	97812e45-0a4d-437b-8641-cf865d76f464	1	2	22	\N	Sunflower Cooking Oil 1.8L	OIL-SUNFLOWER-1.8L	6.000	0.000	0.000	0.000	3	18.95	\N	\N	17.05	130.75	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-22 09:05:40.926135	2026-07-22 09:05:40.926135
2a51d484-abf8-408c-b3ec-c2fff7ea524a	97812e45-0a4d-437b-8641-cf865d76f464	1	3	22	\N	Sunflower Cooking Oil 1.8L	OIL-SUNFLOWER-1.8L	6.000	0.000	0.000	0.000	3	18.95	\N	\N	17.05	130.75	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-22 09:05:40.92936	2026-07-22 09:05:40.92936
9611e2f7-3042-4d9b-ae57-9410ca6c91c7	97812e45-0a4d-437b-8641-cf865d76f464	1	4	3	\N	Nadec Full Cream Milk 2L	NADEC-MILK-FW-2L	6.000	0.000	0.000	0.000	3	14.95	\N	\N	13.45	103.15	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-22 09:05:40.932442	2026-07-22 09:05:40.932442
d22c4375-10da-4542-a60f-33699b2c47b3	97812e45-0a4d-437b-8641-cf865d76f464	1	5	22	\N	Sunflower Cooking Oil 1.8L	OIL-SUNFLOWER-1.8L	7.000	0.000	0.000	0.000	3	18.95	\N	\N	17.05	149.70	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-22 09:05:40.935409	2026-07-22 09:05:40.935409
4d51a356-35a5-4d98-a9fc-7649c517971c	d3bb4057-5f3d-4eaa-aa7f-2dd14ecfeebb	1	1	22	\N	Sunflower Cooking Oil 1.8L	OIL-SUNFLOWER-1.8L	14.000	0.000	0.000	0.000	3	18.95	\N	\N	17.05	282.35	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-22 09:52:19.491342	2026-07-22 09:52:19.491342
59cb0035-152d-438d-9f50-5adcb2cd270b	c2fc4fa8-a714-458d-9656-f82085a9c28b	1	1	22	\N	Sunflower Cooking Oil 1.8L	OIL-SUNFLOWER-1.8L	14.000	0.000	0.000	0.000	3	18.95	\N	\N	17.05	282.35	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-22 09:54:51.079511	2026-07-22 09:54:51.079511
838e7644-3ce0-4416-b82b-a9b411d2822b	c2fc4fa8-a714-458d-9656-f82085a9c28b	1	2	35	\N	Lux Beauty Soap 120g	LUX-SOAP-120G	1.000	0.000	0.000	0.000	10	3.50	\N	\N	0.53	4.03	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-22 09:54:51.083972	2026-07-22 09:54:51.083972
82c01550-a83f-4810-b103-b97b395c740a	dad48529-23c0-4980-8cfc-648f16494e67	1	1	33	\N	Dettol Original Soap 125g	DETTOL-SOAP-125G	1.000	0.000	0.000	0.000	10	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-22 10:01:20.560272	2026-07-22 10:01:20.560272
4558da76-b9ee-4ef4-be21-ce7ea7e03a87	dad48529-23c0-4980-8cfc-648f16494e67	1	2	26	\N	California Garden Baked Beans 400g	CALGARDEN-BEANS-400G	1.000	0.000	0.000	0.000	8	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-22 10:01:20.565562	2026-07-22 10:01:20.565562
32a22366-4209-49ed-ad19-e500629785b8	dad48529-23c0-4980-8cfc-648f16494e67	1	3	6	\N	Almarai Cheese Slices 200g	ALMARAI-CHEESE-SLICE-200G	1.000	0.000	0.000	0.000	10	12.95	\N	\N	1.94	14.89	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-22 10:01:20.56953	2026-07-22 10:01:20.56953
b3860160-4d37-4c85-84ad-7bfa860978ad	98ec425c-3a11-4f1e-ac90-5d96ab6e5819	1	1	33	\N	Dettol Original Soap 125g	DETTOL-SOAP-125G	1.000	0.000	0.000	0.000	10	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-22 10:02:26.793413	2026-07-22 10:02:26.793413
28dc3f6b-b9b7-49f5-8e20-8cf60064e7db	98ec425c-3a11-4f1e-ac90-5d96ab6e5819	1	2	26	\N	California Garden Baked Beans 400g	CALGARDEN-BEANS-400G	1.000	0.000	0.000	0.000	8	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-22 10:02:26.796809	2026-07-22 10:02:26.796809
ab7fbbb5-b72c-41d9-a10f-88d0657af3ec	b5f91a7e-875e-46c5-94d1-c86525cccd5f	1	1	33	\N	Dettol Original Soap 125g	DETTOL-SOAP-125G	1.000	0.000	0.000	0.000	10	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-22 10:02:51.752264	2026-07-22 10:02:51.752264
4b2d61ff-ee32-46aa-b904-83ca73349f46	72138058-5326-4827-a07b-1b92c01446f3	1	1	33	\N	Dettol Original Soap 125g	DETTOL-SOAP-125G	1.000	0.000	0.000	0.000	10	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-22 10:05:41.927857	2026-07-22 10:05:41.927857
7223c42c-99ac-4ec4-befe-97f074ffb049	489e6dac-d79f-417a-8014-cdc3360080f7	1	1	33	\N	Dettol Original Soap 125g	DETTOL-SOAP-125G	1.000	0.000	0.000	0.000	10	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-22 10:07:14.535207	2026-07-22 10:07:14.535207
8fdde528-3787-47f9-b092-10ffce38fb4b	11003230-d7b2-44d1-b057-5a5347af254d	1	1	40	\N	Finish Dishwasher Tablets 40pcs	FINISH-TABS-40PCS	1.000	0.000	0.000	0.000	4	48.50	\N	\N	7.27	55.77	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-22 13:11:05.662225	2026-07-22 13:11:05.662225
72bcc36b-8655-412a-9081-f54ebb3d5bda	11003230-d7b2-44d1-b057-5a5347af254d	1	2	22	\N	Sunflower Cooking Oil 1.8L	OIL-SUNFLOWER-1.8L	6.000	0.000	0.000	0.000	3	18.95	\N	\N	17.05	130.75	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-22 13:11:05.670655	2026-07-22 13:11:05.670655
9e561912-3f5a-496e-89db-84ebf82f48be	11003230-d7b2-44d1-b057-5a5347af254d	1	3	8	\N	Fresh White Eggs 30 Pieces	EGGS-WHITE-30PCS	10.000	0.000	0.000	0.000	1	18.00	\N	\N	27.00	207.00	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-22 13:11:05.673511	2026-07-22 13:11:05.673511
5c44e330-c7be-458f-a091-e75647f46a26	e04934f5-c4d6-4676-9932-8cfa48b76178	1	1	8	\N	Fresh White Eggs 30 Pieces	EGGS-WHITE-30PCS	10.000	0.000	0.000	0.000	1	18.00	\N	\N	27.00	207.00	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-22 13:19:08.047146	2026-07-22 13:19:08.047146
df392de0-e4ce-41e7-8734-5d951a583f32	bb618d80-6189-4df6-91a2-263755459002	1	1	33	\N	Dettol Original Soap 125g	DETTOL-SOAP-125G	1.000	0.000	0.000	0.000	10	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-23 06:18:29.326236	2026-07-23 06:18:29.326236
77eae67e-e5b5-4946-bc10-8012f0df62ae	bb618d80-6189-4df6-91a2-263755459002	1	2	35	\N	Lux Beauty Soap 120g	LUX-SOAP-120G	1.000	0.000	0.000	0.000	10	3.50	\N	\N	0.53	4.03	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-23 06:18:29.334637	2026-07-23 06:18:29.334637
be254388-7573-4d33-a8e9-f0d93dae7b8d	bb618d80-6189-4df6-91a2-263755459002	1	3	6	\N	Almarai Cheese Slices 200g	ALMARAI-CHEESE-SLICE-200G	1.000	0.000	0.000	0.000	10	12.95	\N	\N	1.94	14.89	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-23 06:18:29.337417	2026-07-23 06:18:29.337417
5b2ad169-4023-4d2d-ae62-113687769bca	bb618d80-6189-4df6-91a2-263755459002	1	4	23	\N	Corn Oil 1.8L	OIL-CORN-1.8L	1.000	0.000	0.000	0.000	3	19.95	\N	\N	2.99	22.94	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-23 06:18:29.340658	2026-07-23 06:18:29.340658
ce9ed16e-ae3b-47c0-b308-0f7cd816052d	fc451cb2-50b8-45d3-80a5-e478c795339a	1	1	34	\N	Dove Body Wash 500ml	DOVE-BODYWASH-500ML	1.000	0.000	0.000	0.000	11	24.95	\N	\N	3.74	28.69	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-23 06:28:34.492641	2026-07-23 06:28:34.492641
f93a6f15-bf89-41bc-b4e1-d487e01b45a1	fc451cb2-50b8-45d3-80a5-e478c795339a	1	2	33	\N	Dettol Original Soap 125g	DETTOL-SOAP-125G	1.000	0.000	0.000	0.000	10	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-23 06:28:34.497829	2026-07-23 06:28:34.497829
69af6da6-6e4a-4395-ab5e-8a8e8064f388	fc451cb2-50b8-45d3-80a5-e478c795339a	1	3	35	\N	Lux Beauty Soap 120g	LUX-SOAP-120G	1.000	0.000	0.000	0.000	10	3.50	\N	\N	0.53	4.03	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-23 06:28:34.500809	2026-07-23 06:28:34.500809
ac00b2f7-c21e-446f-9bbb-d82930ffb617	604348ff-bd9a-480f-9d9e-381151944683	1	1	33	\N	Dettol Original Soap 125g	DETTOL-SOAP-125G	1.000	0.000	0.000	0.000	10	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-23 06:37:55.631893	2026-07-23 06:37:55.631893
6d3bd4d3-e662-49c1-bbf1-8081d61e9efd	1c15059e-0602-4345-8a9e-3d7aec6962a4	1	1	26	\N	California Garden Baked Beans 400g	CALGARDEN-BEANS-400G	1.000	0.000	0.000	0.000	8	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-23 08:31:04.312582	2026-07-23 08:31:04.312582
00dd119f-0fdf-4be2-b858-6d690b90ed2e	1c15059e-0602-4345-8a9e-3d7aec6962a4	1	2	35	\N	Lux Beauty Soap 120g	LUX-SOAP-120G	1.000	0.000	0.000	0.000	10	3.50	\N	\N	0.53	4.03	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-23 08:31:04.321125	2026-07-23 08:31:04.321125
d93ed5f8-70a5-4ee4-9868-4fe383f5b131	1c15059e-0602-4345-8a9e-3d7aec6962a4	1	3	27	\N	California Garden Tuna Chunks 185g	CALGARDEN-TUNA-185G	1.000	0.000	0.000	0.000	8	8.95	\N	\N	1.34	10.29	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-23 08:31:04.324196	2026-07-23 08:31:04.324196
02726bf2-c339-485c-a36b-d2490f4b03fb	275b4bb0-20a8-430e-be00-7e0ee053f49f	1	1	33	\N	Dettol Original Soap 125g	DETTOL-SOAP-125G	1.000	0.000	0.000	0.000	10	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-23 10:54:04.972849	2026-07-23 10:54:04.972849
75c669c9-4150-47b6-b976-23845f0fa70b	25f422ca-343b-4dac-b452-96f9e58dbc82	1	1	22	\N	Sunflower Cooking Oil 1.8L	OIL-SUNFLOWER-1.8L	6.000	0.000	0.000	0.000	3	18.95	\N	\N	17.05	130.75	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-23 11:07:35.563214	2026-07-23 11:07:35.563214
29a02373-9c9f-4689-856f-fbfe4941396e	25f422ca-343b-4dac-b452-96f9e58dbc82	1	2	22	\N	Sunflower Cooking Oil 1.8L	OIL-SUNFLOWER-1.8L	6.000	0.000	0.000	0.000	3	18.95	\N	\N	17.05	130.75	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-23 11:07:35.568491	2026-07-23 11:07:35.568491
b369d142-c1a9-4d62-b525-8cf3da80901a	25f422ca-343b-4dac-b452-96f9e58dbc82	1	3	40	\N	Finish Dishwasher Tablets 40pcs	FINISH-TABS-40PCS	6.000	0.000	0.000	0.000	4	48.50	\N	\N	43.65	334.65	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-23 11:07:35.571113	2026-07-23 11:07:35.571113
61c3b086-fa50-4878-807d-d23a3f76a42f	60f69ca5-9235-44c5-bfd6-acb162e8e2c3	1	1	34	\N	Dove Body Wash 500ml	DOVE-BODYWASH-500ML	1.000	0.000	0.000	0.000	11	24.95	\N	\N	3.74	28.69	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-23 11:16:48.147847	2026-07-23 11:16:48.147847
1dc7592a-291f-4112-9a44-f9484a9a0542	60f69ca5-9235-44c5-bfd6-acb162e8e2c3	1	2	26	\N	California Garden Baked Beans 400g	CALGARDEN-BEANS-400G	1.000	0.000	0.000	0.000	8	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-23 11:16:48.157477	2026-07-23 11:16:48.157477
1676583f-3457-4367-bccb-e3bc1d7cfe17	60f69ca5-9235-44c5-bfd6-acb162e8e2c3	1	3	40	\N	Finish Dishwasher Tablets 40pcs	FINISH-TABS-40PCS	1.000	0.000	0.000	0.000	4	48.50	\N	\N	7.27	55.77	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-23 11:16:48.160582	2026-07-23 11:16:48.160582
b37b9784-5df4-4af3-9599-4b77e717baf6	60f69ca5-9235-44c5-bfd6-acb162e8e2c3	1	4	40	\N	Finish Dishwasher Tablets 40pcs	FINISH-TABS-40PCS	3.000	0.000	0.000	0.000	4	48.50	\N	\N	21.82	167.32	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-23 11:16:48.163745	2026-07-23 11:16:48.163745
2d16c051-15c8-4711-9907-6f0aa13598e1	28789b9c-918d-4697-ae9c-19feed267c18	1	1	34	\N	Dove Body Wash 500ml	DOVE-BODYWASH-500ML	1.000	0.000	0.000	0.000	11	24.95	\N	\N	3.74	28.69	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-23 11:24:59.689522	2026-07-23 11:24:59.689522
b4704a02-23db-49c9-892b-f3a60828bc92	28789b9c-918d-4697-ae9c-19feed267c18	1	2	15	\N	Almarai Mixed Fruit Juice 1L	ALMARAI-MIXED-1L	1.000	0.000	0.000	0.000	3	7.95	\N	\N	1.19	9.14	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-23 11:24:59.693941	2026-07-23 11:24:59.693941
4728f027-3c9d-455c-b055-1a65ab75b47f	75b6b802-225e-4481-8ef4-cabb0bbfd07c	1	1	33	\N	Dettol Original Soap 125g	DETTOL-SOAP-125G	1.000	0.000	0.000	0.000	10	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-23 11:42:46.155412	2026-07-23 11:42:46.155412
d16fa332-340e-43a2-807b-f1e8ec8c0d61	2b000f45-99b0-4a8a-b4c7-60dcf39af0b6	1	1	33	\N	Dettol Original Soap 125g	DETTOL-SOAP-125G	2.000	0.000	0.000	0.000	10	4.50	\N	\N	0.67	9.67	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-24 09:14:58.917584	2026-07-24 09:14:58.917584
58a71e00-b707-4ed1-aa7f-b7b91c9919d5	2b000f45-99b0-4a8a-b4c7-60dcf39af0b6	1	2	35	\N	Lux Beauty Soap 120g	LUX-SOAP-120G	1.000	0.000	0.000	0.000	10	3.50	\N	\N	0.53	4.03	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-24 09:14:58.927071	2026-07-24 09:14:58.927071
8816b1b7-3e5f-4f8e-a0f8-0c03fcb55a04	2b000f45-99b0-4a8a-b4c7-60dcf39af0b6	1	3	27	\N	California Garden Tuna Chunks 185g	CALGARDEN-TUNA-185G	2.000	0.000	0.000	0.000	8	8.95	\N	\N	1.34	19.24	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-24 09:14:58.930078	2026-07-24 09:14:58.930078
5dd917b2-2415-4af4-a0f2-07e3fe6b916a	2b000f45-99b0-4a8a-b4c7-60dcf39af0b6	1	4	6	\N	Almarai Cheese Slices 200g	ALMARAI-CHEESE-SLICE-200G	1.000	0.000	0.000	0.000	10	12.95	\N	\N	1.94	14.89	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-24 09:14:58.933355	2026-07-24 09:14:58.933355
9f9e1bda-4264-4bb6-882f-5188177710b6	2b000f45-99b0-4a8a-b4c7-60dcf39af0b6	1	5	23	\N	Corn Oil 1.8L	OIL-CORN-1.8L	1.000	0.000	0.000	0.000	3	19.95	\N	\N	2.99	22.94	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-24 09:14:58.936629	2026-07-24 09:14:58.936629
86efd7c0-43bd-4be8-b080-acb80a55b0ee	17ca1d02-1e68-4055-8370-93d99ef0a469	1	1	34	\N	Dove Body Wash 500ml	DOVE-BODYWASH-500ML	1.000	0.000	0.000	0.000	11	24.95	\N	\N	3.74	28.69	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-27 10:47:45.076747	2026-07-27 10:47:45.076747
059badd2-9457-42b8-836a-153f9a685b5a	2bd4df2f-49f8-409f-aaaa-559039b645fd	1	1	34	\N	Dove Body Wash 500ml	DOVE-BODYWASH-500ML	1.000	0.000	0.000	0.000	11	24.95	\N	\N	3.74	28.69	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-27 11:06:26.245697	2026-07-27 11:06:26.245697
b828cf32-51d1-4542-8890-6f5f626c5fac	2bd4df2f-49f8-409f-aaaa-559039b645fd	1	2	26	\N	California Garden Baked Beans 400g	CALGARDEN-BEANS-400G	1.000	0.000	0.000	0.000	8	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-27 11:06:26.256226	2026-07-27 11:06:26.256226
8aa59ec5-10c7-455f-9351-acd48c8b274a	d11c1e2b-df97-465b-bff1-34b6780ff4fc	1	1	34	\N	Dove Body Wash 500ml	DOVE-BODYWASH-500ML	1.000	0.000	0.000	0.000	11	24.95	\N	\N	3.74	28.69	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-27 11:11:19.929954	2026-07-27 11:11:19.929954
bac5e6cc-1f4a-4192-a4b8-95df92155810	358a5da2-fb5c-496d-b582-a1fd790b4535	1	1	17	\N	Lipton Yellow Label Tea 100 Bags	LIPTON-TEA-100BAG	1.000	0.000	0.000	0.000	4	14.95	\N	\N	2.24	17.19	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-27 11:42:20.354351	2026-07-27 11:42:20.354351
e9a032e2-63fa-4afc-8145-374bd974c44b	358a5da2-fb5c-496d-b582-a1fd790b4535	1	2	29	\N	Table Salt 1kg	SALT-TABLE-1KG	1.000	0.000	0.000	0.000	2	3.00	\N	\N	0.45	3.45	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-27 11:42:20.362929	2026-07-27 11:42:20.362929
b5a1cb2c-e6d4-45f3-942a-92782ca6f820	358a5da2-fb5c-496d-b582-a1fd790b4535	1	3	33	\N	Dettol Original Soap 125g	DETTOL-SOAP-125G	1.000	0.000	0.000	0.000	10	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-27 11:42:20.366366	2026-07-27 11:42:20.366366
42dcf4a1-4b31-4c0e-a002-b60b541ffa91	358a5da2-fb5c-496d-b582-a1fd790b4535	1	4	24	\N	Penne Pasta 500g	PASTA-PENNE-500G	1.000	0.000	0.000	0.000	10	6.50	\N	\N	0.97	7.47	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-27 11:42:20.369583	2026-07-27 11:42:20.369583
72bccde1-fc48-4ff4-ad34-f0f271d2f9c7	358a5da2-fb5c-496d-b582-a1fd790b4535	1	5	35	\N	Lux Beauty Soap 120g	LUX-SOAP-120G	8.000	0.000	0.000	0.000	10	3.50	0.00	0.00	4.24	32.24	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-27 11:42:20.373045	2026-07-27 11:42:20.373045
60731753-2b8d-4d19-a499-20f7f8bda46e	4aa69799-b182-41e0-9fba-92c284472094	1	1	14	\N	Almarai Orange Juice 1L	ALMARAI-ORANGE-1L	6.000	0.000	0.000	0.000	3	7.95	0.00	0.00	7.14	54.84	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-27 11:56:46.14556	2026-07-27 11:56:46.14556
86496162-549c-41d4-b6cb-0ac492c0c685	4aa69799-b182-41e0-9fba-92c284472094	1	2	33	\N	Dettol Original Soap 125g	DETTOL-SOAP-125G	43.000	0.000	0.000	0.000	10	4.50	0.00	0.00	28.81	222.31	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-27 11:56:46.15391	2026-07-27 11:56:46.15391
a5e9434a-63dd-4d5c-bcae-959cf5ba5164	4aa69799-b182-41e0-9fba-92c284472094	1	3	24	\N	Penne Pasta 500g	PASTA-PENNE-500G	1.000	0.000	0.000	0.000	10	6.50	\N	\N	0.97	7.47	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-27 11:56:46.157259	2026-07-27 11:56:46.157259
13aa31f8-d690-411d-94e0-8ed9d507c16e	4aa69799-b182-41e0-9fba-92c284472094	1	4	24	\N	Penne Pasta 500g	PASTA-PENNE-500G	1.000	0.000	0.000	0.000	10	6.50	\N	\N	0.97	7.47	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-27 11:56:46.16053	2026-07-27 11:56:46.16053
f67436b0-8c7c-471d-a716-b6a17f896f73	4aa69799-b182-41e0-9fba-92c284472094	1	5	24	\N	Penne Pasta 500g	PASTA-PENNE-500G	1.000	0.000	0.000	0.000	10	6.50	\N	\N	0.97	7.47	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-27 11:56:46.163758	2026-07-27 11:56:46.163758
c9846bef-de9c-4992-93f1-dd3dccbd820d	4aa69799-b182-41e0-9fba-92c284472094	1	6	29	\N	Table Salt 1kg	SALT-TABLE-1KG	1.000	0.000	0.000	0.000	2	3.00	\N	\N	0.45	3.45	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-27 11:56:46.16723	2026-07-27 11:56:46.16723
f213ea8e-22f7-4072-9013-73a564fe1866	4aa69799-b182-41e0-9fba-92c284472094	1	7	29	\N	Table Salt 1kg	SALT-TABLE-1KG	1.000	0.000	0.000	0.000	2	3.00	\N	\N	0.45	3.45	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-27 11:56:46.170604	2026-07-27 11:56:46.170604
5e1fca8e-e342-4e7b-b88c-1a452af016ec	4aa69799-b182-41e0-9fba-92c284472094	1	8	29	\N	Table Salt 1kg	SALT-TABLE-1KG	1.000	0.000	0.000	0.000	2	3.00	\N	\N	0.45	3.45	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-27 11:56:46.173041	2026-07-27 11:56:46.173041
a60f6c1a-7870-461a-afff-9fdebaf6f01a	f556b2b5-b3c8-4f70-8ccb-cc09f953a861	1	1	29	\N	Table Salt 1kg	SALT-TABLE-1KG	8.000	0.000	0.000	0.000	2	3.00	0.00	0.00	3.60	27.60	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-27 12:05:02.739341	2026-07-27 12:05:02.739341
22de0ac7-6f07-4723-914a-30adeeabf3d5	f556b2b5-b3c8-4f70-8ccb-cc09f953a861	1	2	35	\N	Lux Beauty Soap 120g	LUX-SOAP-120G	39.000	0.000	0.000	0.000	10	3.50	0.00	0.00	20.67	157.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-27 12:05:02.74342	2026-07-27 12:05:02.74342
85701d41-e0e3-4619-8a20-27f0107dc66a	f556b2b5-b3c8-4f70-8ccb-cc09f953a861	1	3	14	\N	Almarai Orange Juice 1L	ALMARAI-ORANGE-1L	55.000	0.000	0.000	0.000	3	7.95	0.00	0.00	65.45	502.70	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-27 12:05:02.745548	2026-07-27 12:05:02.745548
fca0583b-24cf-4e50-9295-f00d1e670ef5	f556b2b5-b3c8-4f70-8ccb-cc09f953a861	1	4	34	\N	Dove Body Wash 500ml	DOVE-BODYWASH-500ML	2.000	0.000	0.000	0.000	11	24.95	0.00	0.00	7.48	57.38	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-27 12:05:02.747893	2026-07-27 12:05:02.747893
1de243a7-84bb-45db-9d9f-a3a307c3e918	f556b2b5-b3c8-4f70-8ccb-cc09f953a861	1	5	17	\N	Lipton Yellow Label Tea 100 Bags	LIPTON-TEA-100BAG	3.000	0.000	0.000	0.000	4	14.95	0.00	0.00	6.72	51.57	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-27 12:05:02.750493	2026-07-27 12:05:02.750493
17779df2-86ed-42a7-a2af-0f70d9b03b2c	f556b2b5-b3c8-4f70-8ccb-cc09f953a861	1	6	33	\N	Dettol Original Soap 125g	DETTOL-SOAP-125G	7.000	0.000	0.000	0.000	10	4.50	0.00	0.00	4.69	36.19	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-27 12:05:02.752932	2026-07-27 12:05:02.752932
75e5e551-83db-44f8-812e-8223afadca4d	e983e5b0-8b05-4ace-95b8-b629a31013ac	1	1	14	\N	Almarai Orange Juice 1L	ALMARAI-ORANGE-1L	5.000	0.000	0.000	0.000	3	7.95	0.00	0.00	5.95	45.70	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-27 12:06:45.381384	2026-07-27 12:06:45.381384
77a28b98-e831-44d0-9760-0242602d0a09	e983e5b0-8b05-4ace-95b8-b629a31013ac	1	2	35	\N	Lux Beauty Soap 120g	LUX-SOAP-120G	8.000	0.000	0.000	0.000	10	3.50	0.00	0.00	4.24	32.24	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-27 12:06:45.391426	2026-07-27 12:06:45.391426
caf8a22d-bd60-4c0b-8619-f20ecbf5cdbb	e983e5b0-8b05-4ace-95b8-b629a31013ac	1	3	33	\N	Dettol Original Soap 125g	DETTOL-SOAP-125G	1.000	0.000	0.000	0.000	10	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-27 12:06:45.394773	2026-07-27 12:06:45.394773
2de0ea9b-dc60-4508-8be6-15cfa27f3616	e983e5b0-8b05-4ace-95b8-b629a31013ac	1	4	34	\N	Dove Body Wash 500ml	DOVE-BODYWASH-500ML	1.000	0.000	0.000	0.000	11	24.95	\N	\N	3.74	28.69	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-27 12:06:45.397895	2026-07-27 12:06:45.397895
7679b75a-e07f-4561-bc75-66b29c4712ff	e983e5b0-8b05-4ace-95b8-b629a31013ac	1	5	26	\N	California Garden Baked Beans 400g	CALGARDEN-BEANS-400G	5.000	0.000	0.000	0.000	8	4.50	0.00	0.00	3.35	25.85	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-27 12:06:45.401502	2026-07-27 12:06:45.401502
f8ee3da4-8c9a-476f-b81d-b1e297e603f6	6fd4ff9a-45d4-4ed7-8002-8b58edc2c9ae	1	1	33	\N	Dettol Original Soap 125g	DETTOL-SOAP-125G	1.000	0.000	0.000	0.000	10	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-28 08:35:54.508952	2026-07-28 08:35:54.508952
a4f1874c-5779-4939-8c63-20f190ac1e50	6fd4ff9a-45d4-4ed7-8002-8b58edc2c9ae	1	2	27	\N	California Garden Tuna Chunks 185g	CALGARDEN-TUNA-185G	1.000	0.000	0.000	0.000	8	8.95	\N	\N	1.34	10.29	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-28 08:35:54.517455	2026-07-28 08:35:54.517455
d7bb72eb-e3ba-4858-9d7b-36e296e0f07d	6fd4ff9a-45d4-4ed7-8002-8b58edc2c9ae	1	3	35	\N	Lux Beauty Soap 120g	LUX-SOAP-120G	1.000	0.000	0.000	0.000	10	3.50	\N	\N	0.53	4.03	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-28 08:35:54.520271	2026-07-28 08:35:54.520271
13e588a2-79e8-4018-8f23-87716fc002d2	6fd4ff9a-45d4-4ed7-8002-8b58edc2c9ae	1	4	6	\N	Almarai Cheese Slices 200g	ALMARAI-CHEESE-SLICE-200G	1.000	0.000	0.000	0.000	10	12.95	\N	\N	1.94	14.89	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-28 08:35:54.52326	2026-07-28 08:35:54.52326
a1b75058-9cfe-4d76-8ce2-d0a750133296	2289e488-034e-4715-803e-93febd00c31b	1	1	33	\N	Dettol Original Soap 125g	DETTOL-SOAP-125G	1.000	0.000	0.000	0.000	10	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-28 08:37:14.658897	2026-07-28 08:37:14.658897
367d9852-ae7b-4d5e-ba7c-7214cb3f07c4	2289e488-034e-4715-803e-93febd00c31b	1	2	35	\N	Lux Beauty Soap 120g	LUX-SOAP-120G	1.000	0.000	0.000	0.000	10	3.50	\N	\N	0.53	4.03	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-28 08:37:14.666542	2026-07-28 08:37:14.666542
5916eedf-ae98-4bce-ac92-5de5d7426987	2289e488-034e-4715-803e-93febd00c31b	1	3	7	\N	Almarai Feta Cheese 400g	ALMARAI-FETA-CHEESE-400G	1.000	0.000	0.000	0.000	10	15.50	\N	\N	2.32	17.82	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-28 08:37:14.669624	2026-07-28 08:37:14.669624
49210164-736e-400f-b21e-f4a9690b59de	2289e488-034e-4715-803e-93febd00c31b	1	4	22	\N	Sunflower Cooking Oil 1.8L	OIL-SUNFLOWER-1.8L	1.000	0.000	0.000	0.000	3	18.95	\N	\N	2.84	21.79	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-28 08:37:14.672764	2026-07-28 08:37:14.672764
a57d2997-0e31-444c-86cb-4c9d311db835	27b796e5-3c0e-40f5-810e-027abd4a7ae4	1	1	22	\N	Sunflower Cooking Oil 1.8L	OIL-SUNFLOWER-1.8L	6.000	0.000	0.000	0.000	3	18.95	\N	\N	17.05	130.75	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 08:02:53.780848	2026-07-30 08:02:53.780848
eb5fe8d7-8e72-490d-92f3-a0f3fdd7dbdd	5fd9d2c3-7306-4627-994e-7fbfd52e50a0	1	1	34	\N	Dove Body Wash 500ml	DOVE-BODYWASH-500ML	1.000	0.000	0.000	0.000	11	24.95	\N	\N	3.74	28.69	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 08:10:35.377004	2026-07-30 08:10:35.377004
eaf38462-483d-4168-ba61-44bdecf2d407	5fd9d2c3-7306-4627-994e-7fbfd52e50a0	1	2	26	\N	California Garden Baked Beans 400g	CALGARDEN-BEANS-400G	1.000	0.000	0.000	0.000	8	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 08:10:35.381567	2026-07-30 08:10:35.381567
b979d04d-d17a-41e2-b79c-0f52ac94db89	5a6f336d-d29c-45b1-955e-e19af7441324	1	1	34	\N	Dove Body Wash 500ml	DOVE-BODYWASH-500ML	3.000	0.000	0.000	0.000	11	24.95	0.00	0.00	11.22	86.07	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 10:59:28.619179	2026-07-30 10:59:28.619179
7a73f786-accc-4a9b-a2f6-7f68e47de16f	5a6f336d-d29c-45b1-955e-e19af7441324	1	2	33	\N	Dettol Original Soap 125g	DETTOL-SOAP-125G	1.000	0.000	0.000	0.000	10	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 10:59:28.628028	2026-07-30 10:59:28.628028
2a27259e-9098-4d66-b3c7-1e8f868c5dfc	a4e64219-9b7f-45c4-9a35-106788960f26	1	1	33	\N	Dettol Original Soap 125g	DETTOL-SOAP-125G	1.000	0.000	0.000	0.000	10	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:04:16.897011	2026-07-30 11:04:16.897011
0d85453a-0c66-467d-af2a-352aae3bc1fd	a4e64219-9b7f-45c4-9a35-106788960f26	1	2	34	\N	Dove Body Wash 500ml	DOVE-BODYWASH-500ML	1.000	0.000	0.000	0.000	11	24.95	\N	\N	3.74	28.69	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:04:16.906626	2026-07-30 11:04:16.906626
d99730f8-3cfd-4c79-a4c6-c86885799d02	a4e64219-9b7f-45c4-9a35-106788960f26	1	3	35	\N	Lux Beauty Soap 120g	LUX-SOAP-120G	1.000	0.000	0.000	0.000	10	3.50	\N	\N	0.53	4.03	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:04:16.909565	2026-07-30 11:04:16.909565
9cffff6b-01d3-4517-a59d-19ac4afc19f3	a4e64219-9b7f-45c4-9a35-106788960f26	1	4	26	\N	California Garden Baked Beans 400g	CALGARDEN-BEANS-400G	1.000	0.000	0.000	0.000	8	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:04:16.912537	2026-07-30 11:04:16.912537
7e818c00-2787-40e9-892a-1860ee3970c7	a4e64219-9b7f-45c4-9a35-106788960f26	1	5	6	\N	Almarai Cheese Slices 200g	ALMARAI-CHEESE-SLICE-200G	1.000	0.000	0.000	0.000	10	12.95	\N	\N	1.94	14.89	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:04:16.915719	2026-07-30 11:04:16.915719
ad8678ad-1ff1-4448-a833-5cd42f59fb17	a4e64219-9b7f-45c4-9a35-106788960f26	1	6	7	\N	Almarai Feta Cheese 400g	ALMARAI-FETA-CHEESE-400G	1.000	0.000	0.000	0.000	10	15.50	\N	\N	2.32	17.82	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:04:16.919066	2026-07-30 11:04:16.919066
089331bc-05b8-4ce9-8b9b-86ebb75bf549	a4e64219-9b7f-45c4-9a35-106788960f26	1	7	23	\N	Corn Oil 1.8L	OIL-CORN-1.8L	1.000	0.000	0.000	0.000	3	19.95	\N	\N	2.99	22.94	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:04:16.922166	2026-07-30 11:04:16.922166
13b5466e-81a0-42ff-8266-644b46c7855f	a4e64219-9b7f-45c4-9a35-106788960f26	1	8	22	\N	Sunflower Cooking Oil 1.8L	OIL-SUNFLOWER-1.8L	1.000	0.000	0.000	0.000	3	18.95	\N	\N	2.84	21.79	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:04:16.924423	2026-07-30 11:04:16.924423
fdf1fd54-1951-427e-8402-934ecdfe0d05	a4e64219-9b7f-45c4-9a35-106788960f26	1	9	40	\N	Finish Dishwasher Tablets 40pcs	FINISH-TABS-40PCS	1.000	0.000	0.000	0.000	4	48.50	\N	\N	7.27	55.77	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:04:16.926767	2026-07-30 11:04:16.926767
33099cdb-8963-4848-b17c-3cdad8c8f017	a4e64219-9b7f-45c4-9a35-106788960f26	1	10	41	\N	Palmolive Dishwashing Liquid 750ml	PALMOLIVE-DISH-750ML	1.000	0.000	0.000	0.000	11	12.95	\N	\N	1.94	14.89	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:04:16.930573	2026-07-30 11:04:16.930573
70438b60-01af-4dd4-8c0d-4ebe267aeb78	a4e64219-9b7f-45c4-9a35-106788960f26	1	11	8	\N	Fresh White Eggs 30 Pieces	EGGS-WHITE-30PCS	1.000	0.000	0.000	0.000	1	18.00	\N	\N	2.70	20.70	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:04:16.933065	2026-07-30 11:04:16.933065
b4f519de-c778-408d-84a6-69d0adf43469	a4e64219-9b7f-45c4-9a35-106788960f26	1	12	2	\N	Almarai Low Fat Milk 1L	ALMARAI-MILK-LF-1L	1.000	0.000	0.000	0.000	3	8.50	\N	\N	1.27	9.78	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:04:16.93515	2026-07-30 11:04:16.93515
d900d213-c385-4f38-bba4-60b0413f9e73	a4e64219-9b7f-45c4-9a35-106788960f26	1	13	1	\N	Almarai Fresh Milk Full Fat 1L	ALMARAI-MILK-FW-1L	1.000	0.000	0.000	0.000	3	8.50	\N	\N	1.27	9.78	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:04:16.937371	2026-07-30 11:04:16.937371
491c0128-28ad-4172-af3e-a35814e5dfa0	a4e64219-9b7f-45c4-9a35-106788960f26	1	14	32	\N	Al-Watania Frozen Chicken 1kg	WATANIA-CHICKEN-1KG	1.000	0.000	0.000	0.000	2	22.00	\N	\N	3.30	25.30	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:04:16.939624	2026-07-30 11:04:16.939624
35271ce0-b5c7-4209-8e75-b615cbf94a4f	a4e64219-9b7f-45c4-9a35-106788960f26	1	15	3	\N	Nadec Full Cream Milk 2L	NADEC-MILK-FW-2L	1.000	0.000	0.000	0.000	3	14.95	\N	\N	2.24	17.19	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:04:16.941683	2026-07-30 11:04:16.941683
f604919b-99c7-4695-b5c7-6ff75f4a3ec5	a4e64219-9b7f-45c4-9a35-106788960f26	1	16	30	\N	Sunbulah French Fries 1kg	SUNBULAH-FRIES-1KG	1.000	0.000	0.000	0.000	2	12.95	\N	\N	1.94	14.89	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:04:16.943773	2026-07-30 11:04:16.943773
a884175a-c38f-44c0-9e6e-b583a23d0c40	a4e64219-9b7f-45c4-9a35-106788960f26	1	17	31	\N	Sunbulah Mixed Vegetables 450g	SUNBULAH-VEGETABLES-450G	1.000	0.000	0.000	0.000	10	9.50	\N	\N	1.43	10.93	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:04:16.946198	2026-07-30 11:04:16.946198
284f17a3-b852-4d75-9678-4d323cdbff53	a4e64219-9b7f-45c4-9a35-106788960f26	1	18	14	\N	Almarai Orange Juice 1L	ALMARAI-ORANGE-1L	1.000	0.000	0.000	0.000	3	7.95	\N	\N	1.19	9.14	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:04:16.948328	2026-07-30 11:04:16.948328
dc305e0f-ee56-4977-9f01-d7301ef22f86	a4e64219-9b7f-45c4-9a35-106788960f26	1	19	15	\N	Almarai Mixed Fruit Juice 1L	ALMARAI-MIXED-1L	1.000	0.000	0.000	0.000	3	7.95	\N	\N	1.19	9.14	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:04:16.950519	2026-07-30 11:04:16.950519
880a2014-f30d-462e-82b7-ae483a25c835	a4e64219-9b7f-45c4-9a35-106788960f26	1	20	38	\N	Ariel Washing Powder 2.5kg	ARIEL-POWDER-2.5KG	1.000	0.000	0.000	0.000	2	35.95	\N	\N	5.39	41.34	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:04:16.952811	2026-07-30 11:04:16.952811
eb61297c-ddcb-457c-882b-6aebbcbae1f8	a4e64219-9b7f-45c4-9a35-106788960f26	1	21	39	\N	Persil Liquid Detergent 3L	PERSIL-LIQUID-3L	1.000	0.000	0.000	0.000	3	42.00	\N	\N	6.30	48.30	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:04:16.954966	2026-07-30 11:04:16.954966
de6f6b7b-f7ab-4651-8f1a-abbf2301bcba	a4e64219-9b7f-45c4-9a35-106788960f26	1	22	36	\N	Palmolive Toothpaste 100ml	PALMOLIVE-TOOTHPASTE-100ML	1.000	0.000	0.000	0.000	11	8.95	\N	\N	1.34	10.29	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:04:16.957231	2026-07-30 11:04:16.957231
9d39afcf-7de8-4fd1-9dfa-382a26ef4f6c	a4e64219-9b7f-45c4-9a35-106788960f26	1	23	37	\N	Tide Washing Powder 3kg	TIDE-POWDER-3KG	1.000	0.000	0.000	0.000	2	39.95	\N	\N	5.99	45.94	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:04:16.959497	2026-07-30 11:04:16.959497
db67e370-d33b-4d2f-953e-a1e600a95c3d	a4e64219-9b7f-45c4-9a35-106788960f26	1	24	25	\N	Spaghetti Pasta 500g	PASTA-SPAGHETTI-500G	1.000	0.000	0.000	0.000	10	6.50	\N	\N	0.97	7.47	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:04:16.961691	2026-07-30 11:04:16.961691
29af91ac-b256-4cc0-b80f-ebea44c0fe52	a4e64219-9b7f-45c4-9a35-106788960f26	1	25	20	\N	Basmati Rice 5kg	RICE-BASMATI-5KG	1.000	0.000	0.000	0.000	2	45.00	\N	\N	6.75	51.75	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:04:16.96385	2026-07-30 11:04:16.96385
c48fe4d0-f363-483d-8291-5f964dcbb995	a4e64219-9b7f-45c4-9a35-106788960f26	1	26	21	\N	Basmati Rice 10kg	RICE-BASMATI-10KG	1.000	0.000	0.000	0.000	2	85.00	\N	\N	12.75	97.75	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:04:16.966227	2026-07-30 11:04:16.966227
d1f4dee4-218f-46ce-8d54-037a2e3b3bf9	a4e64219-9b7f-45c4-9a35-106788960f26	1	27	11	\N	Coca-Cola 2L Bottle	COCA-COLA-2L	1.000	0.000	0.000	0.000	7	6.50	\N	\N	0.97	7.47	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:04:16.968346	2026-07-30 11:04:16.968346
3ec3c2fe-66cd-4dfa-a793-6d948e1e09d8	a4e64219-9b7f-45c4-9a35-106788960f26	1	28	9	\N	Coca-Cola 330ml Can	COCA-COLA-330ML	1.000	0.000	0.000	0.000	8	2.00	\N	\N	0.30	2.30	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:04:16.970614	2026-07-30 11:04:16.970614
384cb23e-ce0e-4472-82c0-7e5331f3c42f	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	1	33	\N	Dettol Original Soap 125g	DETTOL-SOAP-125G	1.000	0.000	0.000	0.000	10	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.541481	2026-07-30 11:08:54.541481
d0f5f849-5266-47b3-b76e-52ae5c20d5f4	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	2	34	\N	Dove Body Wash 500ml	DOVE-BODYWASH-500ML	1.000	0.000	0.000	0.000	11	24.95	\N	\N	3.74	28.69	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.550108	2026-07-30 11:08:54.550108
8917fd7f-717e-49c7-9f19-f3aa2e4b732a	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	3	26	\N	California Garden Baked Beans 400g	CALGARDEN-BEANS-400G	1.000	0.000	0.000	0.000	8	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.552957	2026-07-30 11:08:54.552957
21069664-16d4-456f-ab14-5acffee00fe3	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	4	35	\N	Lux Beauty Soap 120g	LUX-SOAP-120G	1.000	0.000	0.000	0.000	10	3.50	\N	\N	0.53	4.03	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.555775	2026-07-30 11:08:54.555775
049e278a-7f8d-47f2-bcb0-01e872b02e86	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	5	27	\N	California Garden Tuna Chunks 185g	CALGARDEN-TUNA-185G	1.000	0.000	0.000	0.000	8	8.95	\N	\N	1.34	10.29	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.558792	2026-07-30 11:08:54.558792
737dde08-8d0c-4262-a5d8-ee6da5249a99	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	6	6	\N	Almarai Cheese Slices 200g	ALMARAI-CHEESE-SLICE-200G	1.000	0.000	0.000	0.000	10	12.95	\N	\N	1.94	14.89	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.561753	2026-07-30 11:08:54.561753
5e4e3063-b537-4f73-86cd-ffdf566726ce	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	7	23	\N	Corn Oil 1.8L	OIL-CORN-1.8L	1.000	0.000	0.000	0.000	3	19.95	\N	\N	2.99	22.94	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.564857	2026-07-30 11:08:54.564857
5f58e161-360b-4640-bb8b-061648468986	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	8	7	\N	Almarai Feta Cheese 400g	ALMARAI-FETA-CHEESE-400G	1.000	0.000	0.000	0.000	10	15.50	\N	\N	2.32	17.82	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.567133	2026-07-30 11:08:54.567133
bb4de653-6d5f-4cf3-be52-70862402cf11	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	9	22	\N	Sunflower Cooking Oil 1.8L	OIL-SUNFLOWER-1.8L	1.000	0.000	0.000	0.000	3	18.95	\N	\N	2.84	21.79	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.569248	2026-07-30 11:08:54.569248
eb3ea546-9a71-4274-9abd-c98ba7b0a7c1	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	10	40	\N	Finish Dishwasher Tablets 40pcs	FINISH-TABS-40PCS	1.000	0.000	0.000	0.000	4	48.50	\N	\N	7.27	55.77	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.571371	2026-07-30 11:08:54.571371
009d75c0-7257-4851-a5c7-e488a5dade74	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	11	8	\N	Fresh White Eggs 30 Pieces	EGGS-WHITE-30PCS	1.000	0.000	0.000	0.000	1	18.00	\N	\N	2.70	20.70	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.573522	2026-07-30 11:08:54.573522
c51ce895-4af8-42f9-88d8-7057f4ff5c8a	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	12	41	\N	Palmolive Dishwashing Liquid 750ml	PALMOLIVE-DISH-750ML	1.000	0.000	0.000	0.000	11	12.95	\N	\N	1.94	14.89	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.575606	2026-07-30 11:08:54.575606
a79fcb7d-1b83-434a-b0aa-21796c072e7c	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	13	1	\N	Almarai Fresh Milk Full Fat 1L	ALMARAI-MILK-FW-1L	1.000	0.000	0.000	0.000	3	8.50	\N	\N	1.27	9.78	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.577699	2026-07-30 11:08:54.577699
26426806-04df-467f-a382-ee3e16df8e4c	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	14	2	\N	Almarai Low Fat Milk 1L	ALMARAI-MILK-LF-1L	1.000	0.000	0.000	0.000	3	8.50	\N	\N	1.27	9.78	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.579897	2026-07-30 11:08:54.579897
928e9ea3-ddbc-40d6-b046-9c8e4929774f	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	15	32	\N	Al-Watania Frozen Chicken 1kg	WATANIA-CHICKEN-1KG	1.000	0.000	0.000	0.000	2	22.00	\N	\N	3.30	25.30	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.582048	2026-07-30 11:08:54.582048
19b554c8-d8ca-46ce-bdf9-2d22ff40b0e3	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	16	3	\N	Nadec Full Cream Milk 2L	NADEC-MILK-FW-2L	1.000	0.000	0.000	0.000	3	14.95	\N	\N	2.24	17.19	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.584129	2026-07-30 11:08:54.584129
65a4a445-8b20-4d2b-ad3e-24c11097fcdc	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	17	30	\N	Sunbulah French Fries 1kg	SUNBULAH-FRIES-1KG	1.000	0.000	0.000	0.000	2	12.95	\N	\N	1.94	14.89	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.586501	2026-07-30 11:08:54.586501
8cba9e06-2d83-4c6b-a1db-44db06cde508	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	18	31	\N	Sunbulah Mixed Vegetables 450g	SUNBULAH-VEGETABLES-450G	1.000	0.000	0.000	0.000	10	9.50	\N	\N	1.43	10.93	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.588633	2026-07-30 11:08:54.588633
2a495556-3ab7-4985-a173-24b96187a6d9	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	19	14	\N	Almarai Orange Juice 1L	ALMARAI-ORANGE-1L	1.000	0.000	0.000	0.000	3	7.95	\N	\N	1.19	9.14	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.590771	2026-07-30 11:08:54.590771
b8687304-98c5-4355-8a4c-b621bb7212fe	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	20	15	\N	Almarai Mixed Fruit Juice 1L	ALMARAI-MIXED-1L	1.000	0.000	0.000	0.000	3	7.95	\N	\N	1.19	9.14	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.593087	2026-07-30 11:08:54.593087
e2344d7d-d9e6-424a-aa79-9b5d712deed2	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	21	39	\N	Persil Liquid Detergent 3L	PERSIL-LIQUID-3L	1.000	0.000	0.000	0.000	3	42.00	\N	\N	6.30	48.30	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.595234	2026-07-30 11:08:54.595234
c2e7db52-8f44-4d98-a0f9-729d04cffe95	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	22	38	\N	Ariel Washing Powder 2.5kg	ARIEL-POWDER-2.5KG	1.000	0.000	0.000	0.000	2	35.95	\N	\N	5.39	41.34	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.597544	2026-07-30 11:08:54.597544
7665d1b8-5343-466e-88c4-c60a622b9c00	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	23	36	\N	Palmolive Toothpaste 100ml	PALMOLIVE-TOOTHPASTE-100ML	1.000	0.000	0.000	0.000	11	8.95	\N	\N	1.34	10.29	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.599737	2026-07-30 11:08:54.599737
daa2475d-bd25-47e8-9391-ae186ba3df51	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	24	25	\N	Spaghetti Pasta 500g	PASTA-SPAGHETTI-500G	1.000	0.000	0.000	0.000	10	6.50	\N	\N	0.97	7.47	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.601872	2026-07-30 11:08:54.601872
98b48bcf-d29e-4c73-ab0a-7b051c32affa	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	25	24	\N	Penne Pasta 500g	PASTA-PENNE-500G	1.000	0.000	0.000	0.000	10	6.50	\N	\N	0.97	7.47	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.604008	2026-07-30 11:08:54.604008
c755ef48-a864-421c-bbde-508777c469aa	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	26	20	\N	Basmati Rice 5kg	RICE-BASMATI-5KG	1.000	0.000	0.000	0.000	2	45.00	\N	\N	6.75	51.75	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.606333	2026-07-30 11:08:54.606333
119b450c-f641-419b-b23d-e87993a96c18	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	27	11	\N	Coca-Cola 2L Bottle	COCA-COLA-2L	1.000	0.000	0.000	0.000	7	6.50	\N	\N	0.97	7.47	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.608483	2026-07-30 11:08:54.608483
c71e7b1f-7791-4840-bcd1-09e69eb6910b	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	28	29	\N	Table Salt 1kg	SALT-TABLE-1KG	1.000	0.000	0.000	0.000	2	3.00	\N	\N	0.45	3.45	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.610685	2026-07-30 11:08:54.610685
9201e228-d832-4966-9f3b-883ea7e51fc4	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	29	10	\N	Pepsi 330ml Can	PEPSI-330ML	1.000	0.000	0.000	0.000	8	2.00	\N	\N	0.30	2.30	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.612919	2026-07-30 11:08:54.612919
c8375161-8569-4977-a53d-60ee355e3f6e	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	30	17	\N	Lipton Yellow Label Tea 100 Bags	LIPTON-TEA-100BAG	1.000	0.000	0.000	0.000	4	14.95	\N	\N	2.24	17.19	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.615004	2026-07-30 11:08:54.615004
2f879176-9d33-41ca-b783-9ff6bd6e92b5	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	31	19	\N	Nescafe Arabian Coffee 200g	NESCAFE-ARABIAN-200G	1.000	0.000	0.000	0.000	10	32.00	\N	\N	4.80	36.80	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.617093	2026-07-30 11:08:54.617093
96829edd-27e6-4858-894a-a55e9998ebd4	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	32	16	\N	Rabea Tea 100 Bags	RABEA-TEA-100BAG	1.000	0.000	0.000	0.000	4	12.50	\N	\N	1.88	14.38	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.619169	2026-07-30 11:08:54.619169
e21c434f-e190-4627-8c4d-3cc8c6d203c0	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	33	13	\N	Bottled Water 1.5L	WATER-1.5L	1.000	0.000	0.000	0.000	7	1.50	\N	\N	0.22	1.73	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.621258	2026-07-30 11:08:54.621258
e707c558-499a-4d08-ab54-06068f2f7f59	54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	1	34	5	\N	Al-Safi Greek Yogurt 170g	ALSAFI-YOGURT-170G	1.000	0.000	0.000	0.000	10	3.50	\N	\N	0.53	4.03	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-30 11:08:54.623362	2026-07-30 11:08:54.623362
e886b0ef-2f47-4c2a-92a5-6d0ee4fbf368	d7b077c6-d905-4a48-b75d-96acf66acca6	1	1	5	\N	Al-Safi Greek Yogurt 170g	ALSAFI-YOGURT-170G	1.000	0.000	0.000	0.000	10	3.50	\N	\N	0.53	4.03	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-31 09:02:44.042278	2026-07-31 09:02:44.042278
06dffbce-3c70-42d6-ab1b-612438e17d9a	d7b077c6-d905-4a48-b75d-96acf66acca6	1	2	5	\N	Al-Safi Greek Yogurt 170g	ALSAFI-YOGURT-170G	1.000	0.000	0.000	0.000	10	3.50	\N	\N	0.53	4.03	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-31 09:02:44.052183	2026-07-31 09:02:44.052183
8f85143e-b0be-454f-b9eb-3a3f9e405548	d7b077c6-d905-4a48-b75d-96acf66acca6	1	3	8	\N	Fresh White Eggs 30 Pieces	EGGS-WHITE-30PCS	1.000	0.000	0.000	0.000	1	18.00	\N	\N	2.70	20.70	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-31 09:02:44.055243	2026-07-31 09:02:44.055243
cfa84208-696e-43d6-b5f7-f2c2567e7c96	d7b077c6-d905-4a48-b75d-96acf66acca6	1	4	5	\N	Al-Safi Greek Yogurt 170g	ALSAFI-YOGURT-170G	1.000	0.000	0.000	0.000	10	3.50	\N	\N	0.53	4.03	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-31 09:02:44.058468	2026-07-31 09:02:44.058468
8855d957-9652-4729-abd3-d60f307eee99	d7b077c6-d905-4a48-b75d-96acf66acca6	1	5	7	\N	Almarai Feta Cheese 400g	ALMARAI-FETA-CHEESE-400G	1.000	0.000	0.000	0.000	10	15.50	\N	\N	2.32	17.82	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-31 09:02:44.061595	2026-07-31 09:02:44.061595
d3356493-0292-4b3a-bf7f-590a6573b7df	d7b077c6-d905-4a48-b75d-96acf66acca6	1	6	5	\N	Al-Safi Greek Yogurt 170g	ALSAFI-YOGURT-170G	1.000	0.000	0.000	0.000	10	3.50	\N	\N	0.53	4.03	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-31 09:02:44.06466	2026-07-31 09:02:44.06466
8be50779-529a-465f-9c12-9d7a73a4086c	d7b077c6-d905-4a48-b75d-96acf66acca6	1	7	5	\N	Al-Safi Greek Yogurt 170g	ALSAFI-YOGURT-170G	1.000	0.000	0.000	0.000	10	3.50	\N	\N	0.53	4.03	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-31 09:02:44.067823	2026-07-31 09:02:44.067823
ed7d5145-52d4-4b55-9c85-de34b061a4f2	d7b077c6-d905-4a48-b75d-96acf66acca6	1	8	5	\N	Al-Safi Greek Yogurt 170g	ALSAFI-YOGURT-170G	1.000	0.000	0.000	0.000	10	3.50	\N	\N	0.53	4.03	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-31 09:02:44.070053	2026-07-31 09:02:44.070053
af52c279-a3ee-44ac-b31d-142d57ceffe7	d7b077c6-d905-4a48-b75d-96acf66acca6	1	9	5	\N	Al-Safi Greek Yogurt 170g	ALSAFI-YOGURT-170G	1.000	0.000	0.000	0.000	10	3.50	\N	\N	0.53	4.03	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-31 09:02:44.072392	2026-07-31 09:02:44.072392
c10fbda6-83e7-4806-8743-daa1ced8c071	bcf0f68b-1976-4186-9409-13d858a952f6	1	1	5	\N	Al-Safi Greek Yogurt 170g	ALSAFI-YOGURT-170G	1.000	0.000	0.000	0.000	10	3.50	\N	\N	0.53	4.03	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-31 09:05:05.116104	2026-07-31 09:05:05.116104
239d2405-bed6-4ed1-b6d3-abbd12092f67	e6c4a620-4b91-4708-bec4-40e04779346b	1	1	33	\N	Dettol Original Soap 125g	DETTOL-SOAP-125G	3.000	0.000	0.000	0.000	10	4.50	0.00	0.00	2.01	15.51	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-31 09:11:02.471786	2026-07-31 09:11:02.471786
cc919936-5156-4352-95a1-11e25eae1308	e6c4a620-4b91-4708-bec4-40e04779346b	1	2	34	\N	Dove Body Wash 500ml	DOVE-BODYWASH-500ML	1.000	0.000	0.000	0.000	11	24.95	\N	\N	3.74	28.69	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-31 09:11:02.47625	2026-07-31 09:11:02.47625
02237dbc-8078-47e5-8ce6-7ad11759a4ba	e6c4a620-4b91-4708-bec4-40e04779346b	1	3	35	\N	Lux Beauty Soap 120g	LUX-SOAP-120G	1.000	0.000	0.000	0.000	10	3.50	\N	\N	0.53	4.03	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-31 09:11:02.478966	2026-07-31 09:11:02.478966
81482efb-f31e-4ea1-827c-2ab8c09ef8e0	e6c4a620-4b91-4708-bec4-40e04779346b	1	4	26	\N	California Garden Baked Beans 400g	CALGARDEN-BEANS-400G	1.000	0.000	0.000	0.000	8	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-31 09:11:02.481656	2026-07-31 09:11:02.481656
5d63cbc7-1091-43ea-a9e5-31aa0a6a8f56	e6c4a620-4b91-4708-bec4-40e04779346b	1	5	27	\N	California Garden Tuna Chunks 185g	CALGARDEN-TUNA-185G	1.000	0.000	0.000	0.000	8	8.95	\N	\N	1.34	10.29	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-31 09:11:02.484246	2026-07-31 09:11:02.484246
c552ab75-409b-4d48-98b5-aee2a34720e9	e6c4a620-4b91-4708-bec4-40e04779346b	1	6	2	\N	Almarai Low Fat Milk 1L	ALMARAI-MILK-LF-1L	1.000	0.000	0.000	0.000	3	8.50	\N	\N	1.27	9.78	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-31 09:11:02.487002	2026-07-31 09:11:02.487002
146a4b84-a25c-4fd8-8538-b789be1d8b8b	0fcd3efb-e505-45fe-b224-b8010407fd0e	1	1	33	\N	Dettol Original Soap 125g	DETTOL-SOAP-125G	1.000	0.000	0.000	0.000	10	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-31 09:13:36.24426	2026-07-31 09:13:36.24426
a9abf26c-ea84-4588-a9a4-ed220794e9a4	0fcd3efb-e505-45fe-b224-b8010407fd0e	1	2	34	\N	Dove Body Wash 500ml	DOVE-BODYWASH-500ML	1.000	0.000	0.000	0.000	11	24.95	\N	\N	3.74	28.69	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-31 09:13:36.247388	2026-07-31 09:13:36.247388
5c369556-2879-4306-8818-99a8ea0d8574	8076b3b4-166f-457d-8358-ba2fd13a3832	1	1	33	\N	Dettol Original Soap 125g	DETTOL-SOAP-125G	2.000	0.000	0.000	0.000	10	4.50	0.00	0.00	1.34	10.34	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-31 10:19:10.727492	2026-07-31 10:19:10.727492
589dde9b-04d1-471e-bcf8-cb4db7cd0a56	8076b3b4-166f-457d-8358-ba2fd13a3832	1	2	35	\N	Lux Beauty Soap 120g	LUX-SOAP-120G	1.000	0.000	0.000	0.000	10	3.50	\N	\N	0.53	4.03	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-31 10:19:10.735929	2026-07-31 10:19:10.735929
6762a9ee-0690-42e4-870c-7d421e98ce67	8076b3b4-166f-457d-8358-ba2fd13a3832	1	3	26	\N	California Garden Baked Beans 400g	CALGARDEN-BEANS-400G	2.000	0.000	0.000	0.000	8	4.50	0.00	0.00	1.34	10.34	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-31 10:19:10.738843	2026-07-31 10:19:10.738843
aa2d1aad-1bd8-479f-a3fa-ed65e40d8ba5	8076b3b4-166f-457d-8358-ba2fd13a3832	1	4	6	\N	Almarai Cheese Slices 200g	ALMARAI-CHEESE-SLICE-200G	2.000	0.000	0.000	0.000	10	12.95	0.00	0.00	3.88	29.78	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-31 10:19:10.741953	2026-07-31 10:19:10.741953
648cca87-c158-41b4-9c14-542a57e1645d	8076b3b4-166f-457d-8358-ba2fd13a3832	1	5	3	\N	Nadec Full Cream Milk 2L	NADEC-MILK-FW-2L	1.000	0.000	0.000	0.000	3	14.95	\N	\N	2.24	17.19	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-31 10:19:10.745015	2026-07-31 10:19:10.745015
51e64007-67ef-468e-8329-5ae036742ffb	8076b3b4-166f-457d-8358-ba2fd13a3832	1	6	31	\N	Sunbulah Mixed Vegetables 450g	SUNBULAH-VEGETABLES-450G	1.000	0.000	0.000	0.000	10	9.50	\N	\N	1.43	10.93	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-31 10:19:10.747953	2026-07-31 10:19:10.747953
d1ed4f45-feb1-4d67-acc5-1da79255a70d	8076b3b4-166f-457d-8358-ba2fd13a3832	1	7	30	\N	Sunbulah French Fries 1kg	SUNBULAH-FRIES-1KG	2.000	0.000	0.000	0.000	2	12.95	0.00	0.00	3.88	29.78	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-31 10:19:10.751102	2026-07-31 10:19:10.751102
1e2729e1-a7ef-46fd-b0b2-00844936e093	8076b3b4-166f-457d-8358-ba2fd13a3832	1	8	41	\N	Palmolive Dishwashing Liquid 750ml	PALMOLIVE-DISH-750ML	1.000	0.000	0.000	0.000	11	12.95	\N	\N	1.94	14.89	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-31 10:19:10.753317	2026-07-31 10:19:10.753317
0350c319-2184-402c-aac4-a9160e587bce	8076b3b4-166f-457d-8358-ba2fd13a3832	1	9	2	\N	Almarai Low Fat Milk 1L	ALMARAI-MILK-LF-1L	1.000	0.000	0.000	0.000	3	8.50	\N	\N	1.27	9.78	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-31 10:19:10.755587	2026-07-31 10:19:10.755587
ef314698-c881-4b6c-8398-385f942e2f37	8076b3b4-166f-457d-8358-ba2fd13a3832	1	10	37	\N	Tide Washing Powder 3kg	TIDE-POWDER-3KG	2.000	0.000	0.000	0.000	2	39.95	0.00	0.00	11.98	91.88	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-07-31 10:19:10.757762	2026-07-31 10:19:10.757762
0e13ac15-3459-4f35-b2e4-51d47fd7c036	098814e2-f46e-4be6-9730-87857812e82a	1	1	35	\N	Lux Beauty Soap 120g	LUX-SOAP-120G	1.000	0.000	0.000	0.000	10	3.50	\N	\N	0.53	4.03	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-08-01 06:18:24.45555	2026-08-01 06:18:24.45555
52151a2d-fbef-4eb6-bc80-e113d10aa481	098814e2-f46e-4be6-9730-87857812e82a	1	2	22	\N	Sunflower Cooking Oil 1.8L	OIL-SUNFLOWER-1.8L	1.000	0.000	0.000	0.000	3	18.95	\N	\N	2.84	21.79	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-08-01 06:18:24.464519	2026-08-01 06:18:24.464519
83e7f800-e9c6-4962-93e4-e1acc13d0aa4	098814e2-f46e-4be6-9730-87857812e82a	1	3	8	\N	Fresh White Eggs 30 Pieces	EGGS-WHITE-30PCS	1.000	0.000	0.000	0.000	1	18.00	\N	\N	2.70	20.70	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-08-01 06:18:24.468094	2026-08-01 06:18:24.468094
c36ad0d0-1798-47df-be36-55e4de389deb	336b87c3-8860-4889-953b-95145ea1aeab	1	1	35	\N	Lux Beauty Soap 120g	LUX-SOAP-120G	4.000	0.000	0.000	0.000	10	3.50	0.00	0.00	2.12	16.12	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-08-01 06:26:48.168901	2026-08-01 06:26:48.168901
3e66548b-a5f2-44f3-96b8-c909336a72df	a7dcdda8-9253-4201-a387-a0276ab91c39	1	1	34	\N	Dove Body Wash 500ml	DOVE-BODYWASH-500ML	1.000	0.000	0.000	0.000	11	24.95	\N	\N	3.74	28.69	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-08-01 06:47:31.594065	2026-08-01 06:47:31.594065
95d04ad6-96ac-476b-b699-322692aaccba	a7dcdda8-9253-4201-a387-a0276ab91c39	1	2	35	\N	Lux Beauty Soap 120g	LUX-SOAP-120G	1.000	0.000	0.000	0.000	10	3.50	\N	\N	0.53	4.03	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-08-01 06:47:31.602394	2026-08-01 06:47:31.602394
3d811adf-1cb9-4323-923d-d7d8393c3ef9	50388992-d30f-4ee7-929d-ee180ede36a7	1	1	34	\N	Dove Body Wash 500ml	DOVE-BODYWASH-500ML	10.000	0.000	0.000	0.000	11	24.95	0.00	0.00	37.40	286.90	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-08-03 05:31:22.824918	2026-08-03 05:31:22.824918
569ac606-8d30-4e2f-a030-86fb87137d14	50388992-d30f-4ee7-929d-ee180ede36a7	1	2	27	\N	California Garden Tuna Chunks 185g	CALGARDEN-TUNA-185G	1.000	0.000	0.000	0.000	8	8.95	\N	\N	1.34	10.29	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-08-03 05:31:22.834354	2026-08-03 05:31:22.834354
51eedf75-ca64-41ec-b3b2-b52d3c97a275	50388992-d30f-4ee7-929d-ee180ede36a7	1	3	23	\N	Corn Oil 1.8L	OIL-CORN-1.8L	5.000	0.000	0.000	0.000	3	19.95	0.00	0.00	14.95	114.70	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-08-03 05:31:22.837677	2026-08-03 05:31:22.837677
c80bc1c5-4a36-42cc-9bce-557ded221714	ffc609d0-0015-4ced-9d82-025d4f8a7ca3	1	1	6	\N	Almarai Cheese Slices 200g	ALMARAI-CHEESE-SLICE-200G	1.000	0.000	0.000	0.000	10	12.95	\N	\N	1.94	14.89	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-08-03 09:55:00.664779	2026-08-03 09:55:00.664779
64cc7612-6f65-4b39-938d-626af1f90a2a	ffc609d0-0015-4ced-9d82-025d4f8a7ca3	1	2	6	\N	Almarai Cheese Slices 200g	ALMARAI-CHEESE-SLICE-200G	1.000	0.000	0.000	0.000	10	12.95	\N	\N	1.94	14.89	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-08-03 09:55:00.673899	2026-08-03 09:55:00.673899
412a369a-247d-4617-bdf6-5a9bbf024864	ffc609d0-0015-4ced-9d82-025d4f8a7ca3	1	3	6	\N	Almarai Cheese Slices 200g	ALMARAI-CHEESE-SLICE-200G	1.000	0.000	0.000	0.000	10	12.95	\N	\N	1.94	14.89	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-08-03 09:55:00.676925	2026-08-03 09:55:00.676925
62d5d880-cd15-4b6b-bc15-e9e0cfc86cd9	ffc609d0-0015-4ced-9d82-025d4f8a7ca3	1	4	8	\N	Fresh White Eggs 30 Pieces	EGGS-WHITE-30PCS	1.000	0.000	0.000	0.000	1	18.00	\N	\N	2.70	20.70	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-08-03 09:55:00.679692	2026-08-03 09:55:00.679692
10be4e07-ecfb-4ede-960e-cfbf4da1076d	8fd7a740-50df-4b24-b468-6eabcd6ec92d	1	1	33	\N	Dettol Original Soap 125g	DETTOL-SOAP-125G	1.000	0.000	0.000	0.000	10	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-08-03 10:34:21.527587	2026-08-03 10:34:21.527587
f1b26e62-b6ad-4641-aa17-525ee9454a04	8387e582-d524-4c27-8c35-24ec3f9e8d60	1	1	33	\N	Dettol Original Soap 125g	DETTOL-SOAP-125G	1.000	0.000	0.000	0.000	10	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-08-03 10:36:05.372569	2026-08-03 10:36:05.372569
a23dc953-3288-4ab6-9f49-069787816dd3	ce52051e-4a6e-440e-909a-86dd0703926d	1	1	35	\N	Lux Beauty Soap 120g	LUX-SOAP-120G	1.000	0.000	0.000	0.000	10	3.50	\N	\N	0.53	4.03	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-08-03 10:36:47.663662	2026-08-03 10:36:47.663662
e19087e1-0020-4e95-9a91-c178b4d71018	ce52051e-4a6e-440e-909a-86dd0703926d	1	2	33	\N	Dettol Original Soap 125g	DETTOL-SOAP-125G	1.000	0.000	0.000	0.000	10	4.50	\N	\N	0.67	5.17	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-08-03 10:36:47.666995	2026-08-03 10:36:47.666995
c5c3d50f-2359-4dab-b2f1-bbf28de60789	480b985a-df97-4c2b-95c3-bef7d77c0e72	1	1	8	\N	Fresh White Eggs 30 Pieces	EGGS-WHITE-30PCS	1.000	0.000	0.000	0.000	1	18.00	\N	\N	2.70	20.70	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-08-03 11:39:23.341481	2026-08-03 11:39:23.341481
75632adc-5812-468e-b6e3-86845884d7ba	332d8ea2-6299-45fc-b833-ad082299dd72	1	1	5	\N	Al-Safi Greek Yogurt 170g	ALSAFI-YOGURT-170G	1.000	0.000	0.000	0.000	10	3.50	\N	\N	0.53	4.03	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-08-03 11:41:39.523518	2026-08-03 11:41:39.523518
e48eb8d1-cc07-40a3-913a-c324e9406fad	0ad0c821-b76b-4f0c-b149-7edf6df71da9	1	1	35	\N	Lux Beauty Soap 120g	LUX-SOAP-120G	2.000	0.000	0.000	0.000	10	3.50	0.00	0.00	1.06	8.06	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-08-11 11:24:31.135266	2026-08-11 11:24:31.135266
cdcfc1f5-3107-49b8-b323-f2047c76a517	1f34855a-1ff6-4e77-a567-65232d4a4504	1	1	35	\N	Lux Beauty Soap 120g	LUX-SOAP-120G	1.000	0.000	0.000	0.000	10	3.50	\N	\N	0.53	4.03	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-08-11 11:46:45.418351	2026-08-11 11:46:45.418351
6d5f3219-38df-47c5-98d9-25568c1f0ec3	1f34855a-1ff6-4e77-a567-65232d4a4504	1	2	33	\N	Dettol Original Soap 125g	DETTOL-SOAP-125G	2.000	0.000	0.000	0.000	10	4.50	0.00	0.00	1.34	10.34	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-08-11 11:46:45.427169	2026-08-11 11:46:45.427169
6cc11e13-088d-4701-8847-0b5417bf8a98	1f34855a-1ff6-4e77-a567-65232d4a4504	1	3	26	\N	California Garden Baked Beans 400g	CALGARDEN-BEANS-400G	3.000	0.000	0.000	0.000	8	4.50	0.00	0.00	2.01	15.51	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-08-11 11:46:45.429889	2026-08-11 11:46:45.429889
9292b6e4-3ff7-4b23-9192-ca6d15150d9b	1f34855a-1ff6-4e77-a567-65232d4a4504	1	4	7	\N	Almarai Feta Cheese 400g	ALMARAI-FETA-CHEESE-400G	1.000	0.000	0.000	0.000	10	15.50	\N	\N	2.32	17.82	1	15.00	\N	\N	\N	pending	\N	\N	\N	\N	2026-08-11 11:46:45.432996	2026-08-11 11:46:45.432996
\.


--
-- Data for Name: sales_orders; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.sales_orders (id, organization_id, order_number, customer_id, store_id, order_date, delivery_date, status, subtotal, discount_amount, tax_amount, total_amount, price_list_id, created_by, metadata, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: sales_orders_v2; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.sales_orders_v2 (id, order_number, organization_id, store_id, customer_id, customer_name, customer_email, customer_phone, order_type, order_status, payment_status, fulfillment_status, sales_channel, order_source, referral_source, source_cart_id, created_by_user_id, assigned_to_user_id, order_date, confirmed_date, expected_delivery_date, actual_delivery_date, cancelled_date, subtotal, discount_amount, tax_amount, shipping_amount, adjustment_amount, total_amount, paid_amount, refunded_amount, balance_due, coupon_code, discount_codes, promotional_credits, shipping_address, billing_address, shipping_method, shipping_carrier, tracking_number, tracking_url, payment_method, payment_gateway, payment_terms, payment_due_date, pos_terminal_id, cashier_id, is_gift, gift_message, special_instructions, internal_notes, tags, priority, metadata, created_at, updated_at) FROM stdin;
dd2f97c3-8864-4ea1-92d9-fba5fe3326bf	ORD-20260720082131	1	1	1	aku			standard	pending	unpaid	unfulfilled	pos	\N	\N	786d75ea-064f-4ec6-9f1e-15a9cd122adc	4	\N	2026-07-20 08:21:31.507485	\N	\N	\N	\N	188.51	0.00	28.27	0.00	0.00	216.78	0.00	0.00	216.78		{}	0.00	{}	{}	standard	\N	\N	\N	\N	\N	\N	\N	8	1	f	\N		\N	\N	normal	{}	2026-07-20 08:21:31.507485	2026-07-20 08:21:31.552289
1c15059e-0602-4345-8a9e-3d7aec6962a4	ORD-20260723083104	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	ca973e0c-c7d6-4a65-a74e-895089e0527a	4	\N	2026-07-23 08:31:04.30609	\N	\N	\N	\N	16.95	0.00	2.54	0.00	0.00	19.49	20.00	0.00	-0.51		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	9	1	f	\N		\N	\N	normal	{}	2026-07-23 08:31:04.30609	2026-07-23 08:31:13.92435
84b01ab7-92d9-42ba-9741-663c8ecc6b11	ORD-20260720083116	1	1	1	aku			standard	fulfilled	paid	fulfilled	pos	\N	\N	1b665464-53d1-4f00-9076-ccc0b9d8d058	4	\N	2026-07-20 08:31:16.39892	\N	\N	\N	\N	326.45	0.00	48.96	0.00	0.00	375.41	375.41	0.00	0.00		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	8	1	f	\N		\N	\N	normal	{}	2026-07-20 08:31:16.39892	2026-07-20 08:38:08.303301
275b4bb0-20a8-430e-be00-7e0ee053f49f	ORD-20260723105404	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	0b90118f-6987-47c4-bc66-3325bd219b63	4	\N	2026-07-23 10:54:04.965803	\N	\N	\N	\N	4.50	0.00	0.67	0.00	0.00	5.17	6.00	0.00	-0.83		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	9	1	f	\N		\N	\N	normal	{}	2026-07-23 10:54:04.965803	2026-07-23 10:54:37.846937
52a68651-bf78-476f-952a-b29109ba9b2f	ORD-20260721071223	1	1	1	aku	guest@gmail.com	03123456789	standard	fulfilled	paid	fulfilled	pos	\N	\N	ea501529-ce47-4e5f-898a-6c9e2304c995	4	\N	2026-07-21 07:12:23.889965	\N	\N	\N	\N	42.40	0.00	6.35	0.00	0.00	48.75	49.00	0.00	-0.25		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	9	1	f	\N		\N	\N	normal	{}	2026-07-21 07:12:23.889965	2026-07-21 07:12:28.846234
f31d00e6-3197-4290-8783-5cb985915208	ORD-20260720110302	1	1	1	aku			standard	fulfilled	paid	fulfilled	pos	\N	\N	56324b30-f1ce-48c0-aea8-be446dff854c	4	\N	2026-07-20 11:03:02.714912	\N	\N	\N	\N	42.40	0.00	6.35	0.00	0.00	48.75	49.00	0.00	-0.25		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	8	1	f	\N		\N	\N	normal	{}	2026-07-20 11:03:02.714912	2026-07-20 11:40:38.603807
7d954505-fbde-4d37-8ea3-f733e5f87429	ORD-20260722073706	1	1	2	Jane Doe	jane@example.com	+15551234567	standard	fulfilled	paid	fulfilled	pos	\N	\N	def65158-beb8-4300-83d5-fba82480ba3e	4	\N	2026-07-22 07:37:06.207043	\N	\N	\N	\N	37.90	0.00	5.68	0.00	0.00	43.58	44.00	0.00	-0.42		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	9	1	f	\N		\N	\N	normal	{}	2026-07-22 07:37:06.207043	2026-07-22 07:37:11.094294
3b0666c1-9ea5-4698-9c8f-06dc39329f39	ORD-20260720114331	1	1	1	aku			standard	fulfilled	paid	fulfilled	pos	\N	\N	b6933e63-f839-4d53-a4da-501b113d878a	4	\N	2026-07-20 11:43:31.225812	\N	\N	\N	\N	28.45	0.00	4.27	0.00	0.00	32.72	33.00	0.00	-0.28		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	8	1	f	\N		\N	\N	normal	{}	2026-07-20 11:43:31.225812	2026-07-20 11:43:36.106361
9bc4c20b-4fb5-4c28-9104-8552a3195649	ORD-20260721071131	1	1	2	Jane Doe	jane@example.com	+15551234567	standard	pending	unpaid	unfulfilled	pos	\N	\N	3dfb4094-071e-40a5-bda1-c6e2159e0c7c	4	\N	2026-07-21 07:11:31.105103	\N	\N	\N	\N	33.95	0.00	5.08	0.00	0.00	39.03	0.00	0.00	39.03		{}	0.00	{}	{}	standard	\N	\N	\N	\N	\N	\N	\N	9	1	f	\N		\N	\N	normal	{}	2026-07-21 07:11:31.105103	2026-07-21 07:11:31.12512
4898af5e-ea60-4761-af8d-08aef13b9274	ORD-20260721071248	1	1	1	aku	guest@gmail.com	03123456789	standard	fulfilled	paid	fulfilled	pos	\N	\N	37447439-4024-451a-8807-b8cfca010487	4	\N	2026-07-21 07:12:48.86144	\N	\N	\N	\N	303.45	0.00	45.52	0.00	0.00	348.97	500.00	0.00	-151.03		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	9	1	f	\N		\N	\N	normal	{}	2026-07-21 07:12:48.86144	2026-07-21 07:12:55.635739
51ce494c-ab3a-4854-bd8e-f5a2cb82c3a5	ORD-20260722062608	1	1	1	aku	guest@gmail.com	03123456789	standard	fulfilled	paid	fulfilled	pos	\N	\N	2a7a15e0-9167-4524-abbd-bfd138bbaa97	4	\N	2026-07-22 06:26:08.053801	\N	\N	\N	\N	20.95	0.00	3.14	0.00	0.00	24.09	25.00	0.00	-0.91		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	9	1	f	\N		\N	\N	normal	{}	2026-07-22 06:26:08.053801	2026-07-22 06:26:13.472845
80dc64ca-4393-4cf3-897b-2ade3321c8a2	ORD-20260722073739	1	1	2	Jane Doe	jane@example.com	+15551234567	standard	fulfilled	paid	fulfilled	pos	\N	\N	d8d59145-07e8-467c-b016-9555c1af7a88	4	\N	2026-07-22 07:37:39.263321	\N	\N	\N	\N	291.00	0.00	43.65	0.00	0.00	334.65	335.00	0.00	-0.35		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	9	1	f	\N		\N	\N	normal	{}	2026-07-22 07:37:39.263321	2026-07-22 07:37:45.05662
97812e45-0a4d-437b-8641-cf865d76f464	ORD-20260722090540	1	1	1	aku	guest@gmail.com	03123456789	standard	pending	unpaid	unfulfilled	pos	\N	\N	6912df30-1a48-4569-af67-9885dccd1109	4	\N	2026-07-22 09:05:40.908668	\N	\N	\N	\N	563.45	0.00	81.65	0.00	0.00	645.10	0.00	0.00	645.10		{}	0.00	{}	{}	standard	\N	\N	\N	\N	\N	\N	\N	8	1	f	\N		\N	\N	normal	{}	2026-07-22 09:05:40.908668	2026-07-22 09:05:40.935409
e72fdba4-6d6f-4c2a-8337-606bdca5c876	ORD-20260722073822	1	1	1	aku	guest@gmail.com	03123456789	standard	fulfilled	paid	fulfilled	pos	\N	\N	baccf90f-7f2c-4974-b935-5754c1f8f28a	4	\N	2026-07-22 07:38:22.782073	\N	\N	\N	\N	425.00	0.00	63.75	0.00	0.00	488.75	488.75	0.00	0.00		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	9	1	f	\N		\N	\N	normal	{}	2026-07-22 07:38:22.782073	2026-07-22 07:38:47.024263
dad48529-23c0-4980-8cfc-648f16494e67	ORD-20260722100120	1	1	2	Jane Doe	jane@example.com	+15551234567	standard	cancelled	unpaid	unfulfilled	pos	\N	\N	161c21c8-61c7-4d8c-b4a0-3c9f7e760e3e	4	\N	2026-07-22 10:01:20.554113	\N	\N	\N	2026-07-22 10:01:41.83068	21.95	0.00	3.28	0.00	0.00	25.23	0.00	0.00	25.23		{}	0.00	{}	{}	standard	\N	\N	\N	\N	\N	\N	\N	8	1	f	\N		\N	\N	normal	{}	2026-07-22 10:01:20.554113	2026-07-22 10:01:41.83068
d3bb4057-5f3d-4eaa-aa7f-2dd14ecfeebb	ORD-20260722095219	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	cancelled	unpaid	unfulfilled	pos	\N	\N	c03314a4-0319-467c-81c7-32b20fe42417	4	\N	2026-07-22 09:52:19.48296	\N	\N	\N	2026-07-22 09:52:28.356945	265.30	0.00	17.05	0.00	0.00	282.35	0.00	0.00	282.35		{}	0.00	{}	{}	standard	\N	\N	\N	\N	\N	\N	\N	8	1	f	\N		\N	\N	normal	{}	2026-07-22 09:52:19.48296	2026-07-22 09:52:28.356945
b5f91a7e-875e-46c5-94d1-c86525cccd5f	ORD-20260722100251	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	6bbea337-0a12-45a9-afb8-0e8eaeeaf9da	4	\N	2026-07-22 10:02:51.748435	\N	\N	\N	\N	4.50	0.00	0.67	0.00	0.00	5.17	6.00	0.00	-0.83		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	8	1	f	\N		\N	\N	normal	{}	2026-07-22 10:02:51.748435	2026-07-22 10:02:56.568605
c2fc4fa8-a714-458d-9656-f82085a9c28b	ORD-20260722095451	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	c03314a4-0319-467c-81c7-32b20fe42417	4	\N	2026-07-22 09:54:51.07443	\N	\N	\N	\N	268.80	0.00	17.58	0.00	0.00	286.38	287.00	0.00	-0.62		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	8	1	f	\N		\N	\N	normal	{}	2026-07-22 09:54:51.07443	2026-07-22 09:58:06.884572
98ec425c-3a11-4f1e-ac90-5d96ab6e5819	ORD-20260722100226	1	1	2	Jane Doe	jane@example.com	+15551234567	standard	fulfilled	paid	fulfilled	pos	\N	\N	161c21c8-61c7-4d8c-b4a0-3c9f7e760e3e	4	\N	2026-07-22 10:02:26.789285	\N	\N	\N	\N	9.00	0.00	1.34	0.00	0.00	10.34	11.00	0.00	-0.66		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	8	1	f	\N		\N	\N	normal	{}	2026-07-22 10:02:26.789285	2026-07-22 10:02:38.590571
489e6dac-d79f-417a-8014-cdc3360080f7	ORD-20260722100714	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	cancelled	unpaid	unfulfilled	pos	\N	\N	156c1666-36f7-4a09-92e5-22dc40cbeba8	4	\N	2026-07-22 10:07:14.531335	\N	\N	\N	2026-07-22 10:07:29.518491	4.50	0.00	0.67	0.00	0.00	5.17	0.00	0.00	5.17		{}	0.00	{}	{}	standard	\N	\N	\N	\N	\N	\N	\N	8	1	f	\N		\N	\N	normal	{}	2026-07-22 10:07:14.531335	2026-07-22 10:07:29.518491
72138058-5326-4827-a07b-1b92c01446f3	ORD-20260722100541	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	cancelled	unpaid	unfulfilled	pos	\N	\N	156c1666-36f7-4a09-92e5-22dc40cbeba8	4	\N	2026-07-22 10:05:41.923776	\N	\N	\N	2026-07-22 10:05:53.967457	4.50	0.00	0.67	0.00	0.00	5.17	0.00	0.00	5.17		{}	0.00	{}	{}	standard	\N	\N	\N	\N	\N	\N	\N	8	1	f	\N		\N	\N	normal	{}	2026-07-22 10:05:41.923776	2026-07-22 10:05:53.967457
e04934f5-c4d6-4676-9932-8cfa48b76178	ORD-20260722131908	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	cancelled	unpaid	unfulfilled	pos	\N	\N	6f23c0ff-e3eb-4ba6-9996-2225ed1afb14	4	\N	2026-07-22 13:19:08.042946	\N	\N	\N	2026-07-22 13:19:12.230385	180.00	0.00	27.00	0.00	0.00	207.00	0.00	0.00	207.00		{}	0.00	{}	{}	standard	\N	\N	\N	\N	\N	\N	\N	8	1	f	\N		\N	\N	normal	{}	2026-07-22 13:19:08.042946	2026-07-22 13:19:12.230385
11003230-d7b2-44d1-b057-5a5347af254d	ORD-20260722131105	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	f3bc32e2-0d76-4bb6-9ed8-7f02b3021566	4	\N	2026-07-22 13:11:05.655009	\N	\N	\N	\N	342.20	0.00	51.32	0.00	0.00	393.52	394.00	0.00	-0.48		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	8	1	f	\N		\N	\N	normal	{}	2026-07-22 13:11:05.655009	2026-07-22 13:11:10.83785
bb618d80-6189-4df6-91a2-263755459002	ORD-20260723061829	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	4985babb-0bfa-4532-b61e-63a989fa9ee5	4	\N	2026-07-23 06:18:29.318977	\N	\N	\N	\N	40.90	0.00	6.13	0.00	0.00	47.03	48.00	0.00	-0.97		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	9	1	f	\N		\N	\N	normal	{}	2026-07-23 06:18:29.318977	2026-07-23 06:18:34.74545
fc451cb2-50b8-45d3-80a5-e478c795339a	ORD-20260723062834	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	93837ac6-d0d5-4bb6-b6fe-aae9278f91c1	4	\N	2026-07-23 06:28:34.485405	\N	\N	\N	\N	32.95	0.00	4.94	0.00	0.00	37.89	38.00	0.00	-0.11		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	9	1	f	\N		\N	\N	normal	{}	2026-07-23 06:28:34.485405	2026-07-23 06:28:39.288752
604348ff-bd9a-480f-9d9e-381151944683	ORD-20260723063755	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	e3f11cd0-804c-4c79-95e0-7b910cfacb6d	4	\N	2026-07-23 06:37:55.624729	\N	\N	\N	\N	4.50	0.00	0.67	0.00	0.00	5.17	6.00	0.00	-0.83		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	9	1	f	\N		\N	\N	normal	{}	2026-07-23 06:37:55.624729	2026-07-23 06:38:00.850442
bcf0f68b-1976-4186-9409-13d858a952f6	ORD-20260731090505	1	8	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	d397fad3-1a15-4db2-a6db-bec5bf2b9297	4	\N	2026-07-31 09:05:05.107532	\N	\N	\N	\N	3.50	0.00	0.53	0.00	0.00	4.03	5.00	0.00	-0.97		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	12	1	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-07-31 09:05:05.107532	2026-07-31 09:05:23.634657
ce52051e-4a6e-440e-909a-86dd0703926d	ORD-20260803103647	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	9f357046-da05-423c-8753-9c3031e83a44	4	\N	2026-08-03 10:36:47.658755	\N	\N	\N	\N	8.00	0.00	1.20	0.00	0.00	9.20	10.00	0.00	-0.80		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	8	1	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-08-03 10:36:47.658755	2026-08-03 10:36:52.148215
25f422ca-343b-4dac-b452-96f9e58dbc82	ORD-20260723110735	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	a0e2078d-2e1b-48a5-896f-ebc458458dc5	4	\N	2026-07-23 11:07:35.557081	\N	\N	\N	\N	518.40	0.00	77.75	0.00	0.00	596.15	597.00	0.00	-0.85		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	8	1	f	\N		\N	\N	normal	{}	2026-07-23 11:07:35.557081	2026-07-23 11:07:40.233784
2289e488-034e-4715-803e-93febd00c31b	ORD-20260728083714	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	7a73337e-1c3b-4c14-9cd9-9e449851d7ee	4	\N	2026-07-28 08:37:14.651939	\N	\N	\N	\N	42.45	0.00	6.36	0.00	0.00	48.81	49.00	0.00	-0.19		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	9	1	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-07-28 08:37:14.651939	2026-07-28 08:37:18.478419
332d8ea2-6299-45fc-b833-ad082299dd72	ORD-20260803114139	1	8	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	4620d829-c1f1-4885-9686-3bf3796b3cbb	4	\N	2026-08-03 11:41:39.519443	\N	\N	\N	\N	3.50	0.00	0.53	0.00	0.00	4.03	5.00	0.00	-0.97		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	12	1	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-08-03 11:41:39.519443	2026-08-03 11:41:43.488103
5a6f336d-d29c-45b1-955e-e19af7441324	ORD-20260730105928	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	05664a1e-fd9e-4045-81b4-fd5cb0d4110c	4	\N	2026-07-30 10:59:28.61012	\N	\N	\N	\N	79.35	0.00	11.89	0.00	0.00	91.24	92.00	0.00	-0.76		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	8	1	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-07-30 10:59:28.61012	2026-07-30 10:59:32.988679
60f69ca5-9235-44c5-bfd6-acb162e8e2c3	ORD-20260723111648	1	1	1	aku	guest@gmail.com	03123456789	standard	fulfilled	paid	fulfilled	pos	\N	\N	b4ab7b6b-7e3c-43a9-b297-7e4950fd3b96	4	\N	2026-07-23 11:16:48.139645	\N	\N	\N	\N	223.45	0.00	33.50	0.00	0.00	256.95	257.00	0.00	-0.05		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	8	1	f	\N		\N	\N	normal	{}	2026-07-23 11:16:48.139645	2026-07-23 11:17:01.368133
a7dcdda8-9253-4201-a387-a0276ab91c39	ORD-20260801064731	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	de7f9185-027d-4156-bf2c-7b03cd948e88	4	\N	2026-08-01 06:47:31.587683	\N	\N	\N	\N	28.45	0.00	4.27	0.00	0.00	32.72	26.00	0.00	6.72		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	8	2	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-08-01 06:47:31.587683	2026-08-01 06:47:39.259653
54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8	ORD-20260730110854	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	aed43720-0c89-4be2-bec4-cbba48662f56	4	\N	2026-07-30 11:08:54.534423	\N	\N	\N	\N	506.38	0.00	75.89	0.00	0.00	582.27	583.00	0.00	-0.73		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	8	1	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-07-30 11:08:54.534423	2026-07-30 11:08:58.292688
8076b3b4-166f-457d-8358-ba2fd13a3832	ORD-20260731101910	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	3b93a016-e219-4b71-a2a8-9ef8886d2aae	44	\N	2026-07-31 10:19:10.720094	\N	\N	\N	\N	199.11	0.00	29.83	0.00	0.00	228.94	229.00	0.00	-0.06		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	8	6	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-07-31 10:19:10.720094	2026-07-31 10:19:16.186314
ffc609d0-0015-4ced-9d82-025d4f8a7ca3	ORD-20260803095500	1	8	\N				standard	fulfilled	paid	fulfilled	pos	\N	\N	b014c45e-d107-4546-9852-0ceafe1d6282	4	\N	2026-08-03 09:55:00.656239	\N	\N	\N	\N	56.85	0.00	8.52	0.00	0.00	65.37	66.00	0.00	-0.63		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	12	1	f	\N		\N	\N	normal	{"customer_name": "Walk-in Guest"}	2026-08-03 09:55:00.656239	2026-08-03 09:55:10.831067
17ca1d02-1e68-4055-8370-93d99ef0a469	ORD-20260727104745	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	2eb60586-650e-4927-a018-945a436f01a3	4	\N	2026-07-27 10:47:45.066852	\N	\N	\N	\N	24.95	0.00	3.74	0.00	0.00	28.69	29.00	0.00	-0.31		{}	0.00	{}	{}	standard	\N	\N	\N	credit card	pos	\N	\N	8	1	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-07-27 10:47:45.066852	2026-07-27 10:48:00.830681
28789b9c-918d-4697-ae9c-19feed267c18	ORD-20260723112459	1	1	1	aku	guest@gmail.com	03123456789	standard	fulfilled	paid	fulfilled	pos	\N	\N	d8c30107-0eb2-4855-bb6a-51cd53505b5d	4	\N	2026-07-23 11:24:59.685238	\N	\N	\N	\N	32.90	0.00	4.93	0.00	0.00	37.83	38.00	0.00	-0.17		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	8	1	f	\N		\N	\N	normal	{}	2026-07-23 11:24:59.685238	2026-07-23 11:25:10.385199
480b985a-df97-4c2b-95c3-bef7d77c0e72	ORD-20260803113923	1	8	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	73419d52-4182-413f-a4da-14cc77155a65	4	\N	2026-08-03 11:39:23.332308	\N	\N	\N	\N	18.00	0.00	2.70	0.00	0.00	20.70	21.00	0.00	-0.30		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	12	1	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-08-03 11:39:23.332308	2026-08-03 11:39:31.856418
27b796e5-3c0e-40f5-810e-027abd4a7ae4	ORD-20260730080253	1	10	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	62161eda-60af-4495-a576-c6d352e0b040	4	\N	2026-07-30 08:02:53.772947	\N	\N	\N	\N	113.70	0.00	17.05	0.00	0.00	130.75	131.00	0.00	-0.25		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	17	1	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-07-30 08:02:53.772947	2026-07-30 08:02:59.314065
75b6b802-225e-4481-8ef4-cabb0bbfd07c	ORD-20260723114246	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	c269c10a-80e9-41ee-a08b-9c5360c2e4d1	4	\N	2026-07-23 11:42:46.148176	\N	\N	\N	\N	4.50	0.00	0.67	0.00	0.00	5.17	6.00	0.00	-0.83		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	9	1	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-07-23 11:42:46.148176	2026-07-23 11:42:50.044955
358a5da2-fb5c-496d-b582-a1fd790b4535	ORD-20260727114220	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	14edb196-5851-4397-b37d-66322afe2a2d	4	\N	2026-07-27 11:42:20.346292	\N	\N	\N	\N	56.95	0.00	8.57	0.00	0.00	65.52	66.00	0.00	-0.48		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	9	1	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-07-27 11:42:20.346292	2026-07-27 11:42:29.409901
5fd9d2c3-7306-4627-994e-7fbfd52e50a0	ORD-20260730081035	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	252baaaf-5ae9-4d3a-84fe-3e0878ca2b3f	4	\N	2026-07-30 08:10:35.372733	\N	\N	\N	\N	29.45	0.00	4.41	0.00	0.00	33.86	50.00	0.00	-16.14		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	8	1	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-07-30 08:10:35.372733	2026-07-30 08:19:56.098533
2bd4df2f-49f8-409f-aaaa-559039b645fd	ORD-20260727110626	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	c62c4d60-b9e9-4c32-9c09-413b37bed03e	4	\N	2026-07-27 11:06:26.237358	\N	\N	\N	\N	29.45	0.00	4.41	0.00	0.00	33.86	34.00	0.00	-0.14		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	9	1	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-07-27 11:06:26.237358	2026-07-27 11:06:39.759726
2b000f45-99b0-4a8a-b4c7-60dcf39af0b6	ORD-20260724091458	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	bfd00d5b-5d48-491b-ade1-b18de6afcb68	4	\N	2026-07-24 09:14:58.90756	\N	\N	\N	\N	63.30	0.00	7.47	0.00	0.00	70.77	71.00	0.00	-0.23		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	9	1	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-07-24 09:14:58.90756	2026-07-24 09:15:04.033643
a4e64219-9b7f-45c4-9a35-106788960f26	ORD-20260730110416	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	6b98ff85-5ee1-493a-9c4f-a1e47cb12075	4	\N	2026-07-30 11:04:16.88959	\N	\N	\N	\N	548.42	0.00	82.20	0.00	0.00	630.62	631.00	0.00	-0.38		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	8	1	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-07-30 11:04:16.88959	2026-07-30 11:04:24.984511
d7b077c6-d905-4a48-b75d-96acf66acca6	ORD-20260731090244	1	8	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	a6a8a674-e57e-4809-9e86-b7d517ef7d4f	4	\N	2026-07-31 09:02:44.031443	\N	\N	\N	\N	58.00	0.00	8.73	0.00	0.00	66.73	67.00	0.00	-0.27		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	12	1	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-07-31 09:02:44.031443	2026-07-31 09:02:52.507259
e6c4a620-4b91-4708-bec4-40e04779346b	ORD-20260731091102	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	8f1bca0d-e6d0-434e-9895-8a68343bd13f	4	\N	2026-07-31 09:11:02.466443	\N	\N	\N	\N	63.91	0.00	9.56	0.00	0.00	73.47	74.00	0.00	-0.53		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	8	1	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-07-31 09:11:02.466443	2026-07-31 09:11:07.943889
d11c1e2b-df97-465b-bff1-34b6780ff4fc	ORD-20260727111119	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	a2eebeb3-662c-46df-82ef-e6ad81058c5b	4	\N	2026-07-27 11:11:19.925869	\N	\N	\N	\N	24.95	0.00	3.74	0.00	0.00	28.69	29.00	0.00	-0.31		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	9	1	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-07-27 11:11:19.925869	2026-07-27 11:11:26.155059
0fcd3efb-e505-45fe-b224-b8010407fd0e	ORD-20260731091336	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	3469e4f7-1aef-475b-951b-2cd3ec962029	4	\N	2026-07-31 09:13:36.240484	\N	\N	\N	\N	29.45	0.00	4.41	0.00	0.00	33.86	34.00	0.00	-0.14		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	8	1	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-07-31 09:13:36.240484	2026-07-31 09:13:42.24131
e983e5b0-8b05-4ace-95b8-b629a31013ac	ORD-20260727120645	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	339d0029-bbdf-49c3-8fc1-cadba066f020	4	\N	2026-07-27 12:06:45.369439	\N	\N	\N	\N	119.70	0.00	17.95	0.00	0.00	137.65	138.00	0.00	-0.35		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	9	1	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-07-27 12:06:45.369439	2026-07-27 12:06:50.029317
098814e2-f46e-4be6-9730-87857812e82a	ORD-20260801061824	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	c37191dd-8749-477b-bbe3-402c2e09a31d	4	\N	2026-08-01 06:18:24.44564	\N	\N	\N	\N	40.45	0.00	6.07	0.00	0.00	46.52	47.00	0.00	-0.48		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	8	2	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-08-01 06:18:24.44564	2026-08-01 06:18:28.175551
f556b2b5-b3c8-4f70-8ccb-cc09f953a861	ORD-20260727120502	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	334043fe-9a22-4e2d-b316-8b780d6717dd	4	\N	2026-07-27 12:05:02.734041	\N	\N	\N	\N	724.00	0.00	108.61	0.00	0.00	832.61	833.00	0.00	-0.39		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	9	1	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-07-27 12:05:02.734041	2026-07-27 12:05:08.876724
4aa69799-b182-41e0-9fba-92c284472094	ORD-20260727115646	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	e1dad131-4aa2-487a-b854-e002b43febe4	4	\N	2026-07-27 11:56:46.137556	\N	\N	\N	\N	269.70	0.00	40.21	0.00	0.00	309.91	310.00	0.00	-0.09		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	9	1	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-07-27 11:56:46.137556	2026-07-27 11:56:51.0094
336b87c3-8860-4889-953b-95145ea1aeab	ORD-20260801062648	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	b627c8a2-bc74-4bb8-9b28-0790aab1266a	4	\N	2026-08-01 06:26:48.164436	\N	\N	\N	\N	14.00	0.00	2.12	0.00	0.00	16.12	17.00	0.00	-0.88		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	8	2	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-08-01 06:26:48.164436	2026-08-01 06:26:52.332477
50388992-d30f-4ee7-929d-ee180ede36a7	ORD-20260803053122	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	7750e699-09d6-45b6-a904-e408f9dabaef	4	\N	2026-08-03 05:31:22.813367	\N	\N	\N	\N	358.20	0.00	53.69	0.00	0.00	411.89	412.00	0.00	-0.11		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	8	1	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-08-03 05:31:22.813367	2026-08-03 05:31:30.355697
8fd7a740-50df-4b24-b468-6eabcd6ec92d	ORD-20260803103421	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	35aa084d-d355-48f6-88d7-8817ee31408b	4	\N	2026-08-03 10:34:21.520275	\N	\N	\N	\N	4.50	0.00	0.67	0.00	0.00	5.17	6.00	0.00	-0.83		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	8	1	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-08-03 10:34:21.520275	2026-08-03 10:34:25.560444
6fd4ff9a-45d4-4ed7-8002-8b58edc2c9ae	ORD-20260728083554	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	25f07dc1-0c1b-4fc4-a9a5-da5398c129a7	4	\N	2026-07-28 08:35:54.498196	\N	\N	\N	\N	29.90	0.00	4.48	0.00	0.00	34.38	35.00	0.00	-0.62		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	9	1	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-07-28 08:35:54.498196	2026-07-28 08:35:58.826576
8387e582-d524-4c27-8c35-24ec3f9e8d60	ORD-20260803103605	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	dabc964f-45e9-4234-a23d-4e2fddcdaaee	4	\N	2026-08-03 10:36:05.368178	\N	\N	\N	\N	4.50	0.00	0.67	0.00	0.00	5.17	6.00	0.00	-0.83		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	8	1	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-08-03 10:36:05.368178	2026-08-03 10:36:09.241354
0ad0c821-b76b-4f0c-b149-7edf6df71da9	ORD-20260811112431	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	a0f2df8d-1907-41dd-927d-749d3e09730b	4	\N	2026-08-11 11:24:31.117197	\N	\N	\N	\N	7.00	0.00	1.06	0.00	0.00	8.06	9.00	0.00	-0.94		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	8	1	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-08-11 11:24:31.117197	2026-08-11 11:24:35.342384
1f34855a-1ff6-4e77-a567-65232d4a4504	ORD-20260811114645	1	1	10	Guest ضيف	Guest@gmail.com	12345678	standard	fulfilled	paid	fulfilled	pos	\N	\N	ecaed928-ee99-437d-8e8d-e46d4597aff8	4	\N	2026-08-11 11:46:45.409218	\N	\N	\N	\N	41.50	0.00	6.20	0.00	0.00	47.70	48.00	0.00	-0.30		{}	0.00	{}	{}	standard	\N	\N	\N	cash	pos	\N	\N	8	1	f	\N		\N	\N	normal	{"customer_name": "Guest ضيف"}	2026-08-11 11:46:45.409218	2026-08-11 11:46:54.854467
\.


--
-- Data for Name: sales_return_lines; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.sales_return_lines (id, return_id, product_id, product_variant_id, original_line_id, quantity, unit_price, refund_amount, return_to_stock, serial_number, batch_number, condition, line_number, metadata, created_at) FROM stdin;
\.


--
-- Data for Name: sales_returns; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.sales_returns (id, return_number, store_id, cashier_id, cashier_session_id, customer_id, original_transaction_id, return_date, return_reason, status, subtotal, tax_amount, total_refund_amount, refund_method, refund_reference, approved_by, notes, metadata, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: stock_count_lines; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.stock_count_lines (id, stock_count_id, product_id, product_variant_id, storage_location_id, expected_quantity, system_quantity, counted_quantity, variance, variance_value, counted_at, uom_id, batch_number, serial_number, metadata, created_at) FROM stdin;
\.


--
-- Data for Name: stock_counts; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.stock_counts (id, count_number, store_id, storage_location_id, count_type, status, scheduled_date, started_at, completed_at, counted_by, approved_by, metadata, created_at) FROM stdin;
\.


--
-- Data for Name: stock_movements; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.stock_movements (id, movement_type, reference_type, reference_id, product_id, product_variant_id, from_store_id, to_store_id, from_location_id, to_location_id, quantity, uom_id, batch_number, serial_number, movement_date, posted_by, status, cost_per_unit, total_value, metadata, created_at) FROM stdin;
1	allocation	sales_order	\N	8	\N	1	\N	\N	\N	10.000	1	\N	\N	2026-07-20 08:21:31.529725	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "5305e57f-4336-4fc7-ade3-40f8013942a1", "sales_order_id": "dd2f97c3-8864-4ea1-92d9-fba5fe3326bf"}	2026-07-20 08:21:31.529725
2	allocation	sales_order	\N	2	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-20 08:21:31.552289	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "078637d7-9c70-48e9-8498-e1e1c4a28fb2", "sales_order_id": "dd2f97c3-8864-4ea1-92d9-fba5fe3326bf"}	2026-07-20 08:21:31.552289
3	allocation	sales_order	\N	40	\N	1	\N	\N	\N	6.000	4	\N	\N	2026-07-20 08:31:16.405848	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "2ab5c658-f421-4c98-a9aa-bbd9da6d0517", "sales_order_id": "84b01ab7-92d9-42ba-9741-663c8ecc6b11"}	2026-07-20 08:31:16.405848
4	allocation	sales_order	\N	23	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-20 08:31:16.414537	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "e581342e-ae81-4604-95cc-7b09f9a8d6d7", "sales_order_id": "84b01ab7-92d9-42ba-9741-663c8ecc6b11"}	2026-07-20 08:31:16.414537
5	allocation	sales_order	\N	7	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-20 08:31:16.417673	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "61c15c47-971e-4fba-90fa-1c11b63b8e7b", "sales_order_id": "84b01ab7-92d9-42ba-9741-663c8ecc6b11"}	2026-07-20 08:31:16.417673
6	sale	sales_order	\N	40	\N	1	\N	\N	\N	6.000	4	\N	\N	2026-07-20 08:38:08.303301	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "2ab5c658-f421-4c98-a9aa-bbd9da6d0517", "sales_order_id": "84b01ab7-92d9-42ba-9741-663c8ecc6b11", "sales_order_number": "ORD-20260720083116"}	2026-07-20 08:38:08.303301
7	sale	sales_order	\N	23	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-20 08:38:08.303301	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "e581342e-ae81-4604-95cc-7b09f9a8d6d7", "sales_order_id": "84b01ab7-92d9-42ba-9741-663c8ecc6b11", "sales_order_number": "ORD-20260720083116"}	2026-07-20 08:38:08.303301
8	sale	sales_order	\N	7	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-20 08:38:08.303301	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "61c15c47-971e-4fba-90fa-1c11b63b8e7b", "sales_order_id": "84b01ab7-92d9-42ba-9741-663c8ecc6b11", "sales_order_number": "ORD-20260720083116"}	2026-07-20 08:38:08.303301
9	allocation	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-20 11:03:02.722886	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "a31e1c0a-fb4f-4152-b6b9-03ae6933dadd", "sales_order_id": "f31d00e6-3197-4290-8783-5cb985915208"}	2026-07-20 11:03:02.722886
10	allocation	sales_order	\N	26	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-20 11:03:02.731844	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "f950c93a-3666-45de-8e22-b772c447276c", "sales_order_id": "f31d00e6-3197-4290-8783-5cb985915208"}	2026-07-20 11:03:02.731844
11	allocation	sales_order	\N	6	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-20 11:03:02.735033	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "8621840c-a049-4c8a-8cbc-6e1d445fa987", "sales_order_id": "f31d00e6-3197-4290-8783-5cb985915208"}	2026-07-20 11:03:02.735033
12	sale	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-20 11:40:38.603807	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "a31e1c0a-fb4f-4152-b6b9-03ae6933dadd", "sales_order_id": "f31d00e6-3197-4290-8783-5cb985915208", "sales_order_number": "ORD-20260720110302"}	2026-07-20 11:40:38.603807
13	sale	sales_order	\N	26	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-20 11:40:38.603807	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "f950c93a-3666-45de-8e22-b772c447276c", "sales_order_id": "f31d00e6-3197-4290-8783-5cb985915208", "sales_order_number": "ORD-20260720110302"}	2026-07-20 11:40:38.603807
14	sale	sales_order	\N	6	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-20 11:40:38.603807	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "8621840c-a049-4c8a-8cbc-6e1d445fa987", "sales_order_id": "f31d00e6-3197-4290-8783-5cb985915208", "sales_order_number": "ORD-20260720110302"}	2026-07-20 11:40:38.603807
15	allocation	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-20 11:43:31.229598	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "128759b0-21b8-47c6-8d19-b5723e2013c7", "sales_order_id": "3b0666c1-9ea5-4698-9c8f-06dc39329f39"}	2026-07-20 11:43:31.229598
16	allocation	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-20 11:43:31.233122	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "3f7c9833-ad5a-4f2f-bf28-77ec34735804", "sales_order_id": "3b0666c1-9ea5-4698-9c8f-06dc39329f39"}	2026-07-20 11:43:31.233122
17	sale	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-20 11:43:36.106361	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "128759b0-21b8-47c6-8d19-b5723e2013c7", "sales_order_id": "3b0666c1-9ea5-4698-9c8f-06dc39329f39", "sales_order_number": "ORD-20260720114331"}	2026-07-20 11:43:36.106361
18	sale	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-20 11:43:36.106361	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "3f7c9833-ad5a-4f2f-bf28-77ec34735804", "sales_order_id": "3b0666c1-9ea5-4698-9c8f-06dc39329f39", "sales_order_number": "ORD-20260720114331"}	2026-07-20 11:43:36.106361
19	allocation	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-21 07:11:31.112697	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "a35a23bd-cc19-4358-8285-ed01321bea9f", "sales_order_id": "9bc4c20b-4fb5-4c28-9104-8552a3195649"}	2026-07-21 07:11:31.112697
20	allocation	sales_order	\N	26	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-21 07:11:31.121976	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "c5a56c3a-6717-4d1b-8963-8785c9bf1ccc", "sales_order_id": "9bc4c20b-4fb5-4c28-9104-8552a3195649"}	2026-07-21 07:11:31.121976
21	allocation	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-21 07:11:31.12512	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "48ce660b-fffd-4994-ae5a-36849b74b7db", "sales_order_id": "9bc4c20b-4fb5-4c28-9104-8552a3195649"}	2026-07-21 07:11:31.12512
22	allocation	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-21 07:12:23.897262	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "ec61e58c-227f-4483-a7a6-f1a0f992fb02", "sales_order_id": "52a68651-bf78-476f-952a-b29109ba9b2f"}	2026-07-21 07:12:23.897262
23	allocation	sales_order	\N	26	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-21 07:12:23.905909	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "effc33fd-a42f-4b25-bd55-b1c5d6637759", "sales_order_id": "52a68651-bf78-476f-952a-b29109ba9b2f"}	2026-07-21 07:12:23.905909
24	allocation	sales_order	\N	6	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-21 07:12:23.909058	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "f2c98948-9f3c-40ed-953d-299fab34cc27", "sales_order_id": "52a68651-bf78-476f-952a-b29109ba9b2f"}	2026-07-21 07:12:23.909058
25	sale	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-21 07:12:28.846234	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "ec61e58c-227f-4483-a7a6-f1a0f992fb02", "sales_order_id": "52a68651-bf78-476f-952a-b29109ba9b2f", "sales_order_number": "ORD-20260721071223"}	2026-07-21 07:12:28.846234
26	sale	sales_order	\N	26	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-21 07:12:28.846234	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "effc33fd-a42f-4b25-bd55-b1c5d6637759", "sales_order_id": "52a68651-bf78-476f-952a-b29109ba9b2f", "sales_order_number": "ORD-20260721071223"}	2026-07-21 07:12:28.846234
27	sale	sales_order	\N	6	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-21 07:12:28.846234	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "f2c98948-9f3c-40ed-953d-299fab34cc27", "sales_order_id": "52a68651-bf78-476f-952a-b29109ba9b2f", "sales_order_number": "ORD-20260721071223"}	2026-07-21 07:12:28.846234
28	allocation	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-21 07:12:48.867624	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "6eba69ac-8f5f-4112-b13e-69ffca7760a7", "sales_order_id": "4898af5e-ea60-4761-af8d-08aef13b9274"}	2026-07-21 07:12:48.867624
455	transfer_out	transfer_request	14	9	\N	4	1	9	2	2.000	5	\N	\N	2026-08-11 07:32:50.47886	1	shipped	\N	\N	{"transfer_number": "TRF-COKE-00454"}	2026-08-11 07:32:50.47886
29	allocation	sales_order	\N	27	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-21 07:12:48.875032	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "7f5623ef-9940-4873-b398-9d497b18ff33", "sales_order_id": "4898af5e-ea60-4761-af8d-08aef13b9274"}	2026-07-21 07:12:48.875032
30	allocation	sales_order	\N	40	\N	1	\N	\N	\N	6.000	4	\N	\N	2026-07-21 07:12:48.878174	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "26913bca-b300-4029-ab98-89ae230e2c5b", "sales_order_id": "4898af5e-ea60-4761-af8d-08aef13b9274"}	2026-07-21 07:12:48.878174
31	sale	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-21 07:12:55.635739	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "6eba69ac-8f5f-4112-b13e-69ffca7760a7", "sales_order_id": "4898af5e-ea60-4761-af8d-08aef13b9274", "sales_order_number": "ORD-20260721071248"}	2026-07-21 07:12:55.635739
32	sale	sales_order	\N	27	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-21 07:12:55.635739	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "7f5623ef-9940-4873-b398-9d497b18ff33", "sales_order_id": "4898af5e-ea60-4761-af8d-08aef13b9274", "sales_order_number": "ORD-20260721071248"}	2026-07-21 07:12:55.635739
33	sale	sales_order	\N	40	\N	1	\N	\N	\N	6.000	4	\N	\N	2026-07-21 07:12:55.635739	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "26913bca-b300-4029-ab98-89ae230e2c5b", "sales_order_id": "4898af5e-ea60-4761-af8d-08aef13b9274", "sales_order_number": "ORD-20260721071248"}	2026-07-21 07:12:55.635739
37	Transfer	Transfer	202300942	33	\N	4	4	10	8	200.000	10	\N	\N	2026-07-22 06:13:09.154	1	Pending	4.5000	900.00	{"remarks": "", "origin_store": "Qitaf Warehouse", "origin_location": "WH-ZONE-C - Zone C - Cold Storage", "destination_store": "Qitaf Warehouse", "destination_location": "WH-ZONE-A - Zone A - Dry Goods"}	2026-07-22 06:13:44.276602
38	allocation	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-22 06:26:08.060627	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "b7160c56-7aed-4a14-9ba1-d40bc1af1f76", "sales_order_id": "51ce494c-ab3a-4854-bd8e-f5a2cb82c3a5"}	2026-07-22 06:26:08.060627
39	allocation	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-22 06:26:08.067873	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "c74d65a2-4e32-4a03-bc00-48080d847eae", "sales_order_id": "51ce494c-ab3a-4854-bd8e-f5a2cb82c3a5"}	2026-07-22 06:26:08.067873
40	allocation	sales_order	\N	6	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-22 06:26:08.070729	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "3d6935c4-d1f5-4e46-8cb9-474e363b3133", "sales_order_id": "51ce494c-ab3a-4854-bd8e-f5a2cb82c3a5"}	2026-07-22 06:26:08.070729
41	sale	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-22 06:26:13.472845	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "b7160c56-7aed-4a14-9ba1-d40bc1af1f76", "sales_order_id": "51ce494c-ab3a-4854-bd8e-f5a2cb82c3a5", "sales_order_number": "ORD-20260722062608"}	2026-07-22 06:26:13.472845
42	sale	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-22 06:26:13.472845	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "c74d65a2-4e32-4a03-bc00-48080d847eae", "sales_order_id": "51ce494c-ab3a-4854-bd8e-f5a2cb82c3a5", "sales_order_number": "ORD-20260722062608"}	2026-07-22 06:26:13.472845
43	sale	sales_order	\N	6	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-22 06:26:13.472845	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "3d6935c4-d1f5-4e46-8cb9-474e363b3133", "sales_order_id": "51ce494c-ab3a-4854-bd8e-f5a2cb82c3a5", "sales_order_number": "ORD-20260722062608"}	2026-07-22 06:26:13.472845
44	allocation	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-22 07:37:06.213954	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "c71ba242-75aa-4836-a144-ed7e00704697", "sales_order_id": "7d954505-fbde-4d37-8ea3-f733e5f87429"}	2026-07-22 07:37:06.213954
45	allocation	sales_order	\N	6	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-22 07:37:06.222227	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "0ea47ebb-12fb-4b38-ae13-9428b227a150", "sales_order_id": "7d954505-fbde-4d37-8ea3-f733e5f87429"}	2026-07-22 07:37:06.222227
46	sale	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-22 07:37:11.094294	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "c71ba242-75aa-4836-a144-ed7e00704697", "sales_order_id": "7d954505-fbde-4d37-8ea3-f733e5f87429", "sales_order_number": "ORD-20260722073706"}	2026-07-22 07:37:11.094294
47	sale	sales_order	\N	6	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-22 07:37:11.094294	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "0ea47ebb-12fb-4b38-ae13-9428b227a150", "sales_order_id": "7d954505-fbde-4d37-8ea3-f733e5f87429", "sales_order_number": "ORD-20260722073706"}	2026-07-22 07:37:11.094294
48	allocation	sales_order	\N	40	\N	1	\N	\N	\N	6.000	4	\N	\N	2026-07-22 07:37:39.271677	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "5eac0b7f-e0f8-4793-bdf7-92dfa3612cce", "sales_order_id": "80dc64ca-4393-4cf3-897b-2ade3321c8a2"}	2026-07-22 07:37:39.271677
49	sale	sales_order	\N	40	\N	1	\N	\N	\N	6.000	4	\N	\N	2026-07-22 07:37:45.05662	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "5eac0b7f-e0f8-4793-bdf7-92dfa3612cce", "sales_order_id": "80dc64ca-4393-4cf3-897b-2ade3321c8a2", "sales_order_number": "ORD-20260722073739"}	2026-07-22 07:37:45.05662
50	allocation	sales_order	\N	21	\N	1	\N	\N	\N	5.000	2	\N	\N	2026-07-22 07:38:22.787444	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "6b76b796-abf4-402a-b104-f5b242bfe4cb", "sales_order_id": "e72fdba4-6d6f-4c2a-8337-606bdca5c876"}	2026-07-22 07:38:22.787444
51	sale	sales_order	\N	21	\N	1	\N	\N	\N	5.000	2	\N	\N	2026-07-22 07:38:47.024263	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "6b76b796-abf4-402a-b104-f5b242bfe4cb", "sales_order_id": "e72fdba4-6d6f-4c2a-8337-606bdca5c876", "sales_order_number": "ORD-20260722073822"}	2026-07-22 07:38:47.024263
52	allocation	sales_order	\N	22	\N	1	\N	\N	\N	6.000	3	\N	\N	2026-07-22 09:05:40.916678	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "6b49bc51-3281-43f6-b01f-c2d34619b8d6", "sales_order_id": "97812e45-0a4d-437b-8641-cf865d76f464"}	2026-07-22 09:05:40.916678
53	allocation	sales_order	\N	22	\N	1	\N	\N	\N	6.000	3	\N	\N	2026-07-22 09:05:40.926135	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "40db0066-fce5-49a7-b9c6-caf037c50722", "sales_order_id": "97812e45-0a4d-437b-8641-cf865d76f464"}	2026-07-22 09:05:40.926135
54	allocation	sales_order	\N	22	\N	1	\N	\N	\N	6.000	3	\N	\N	2026-07-22 09:05:40.92936	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "2a51d484-abf8-408c-b3ec-c2fff7ea524a", "sales_order_id": "97812e45-0a4d-437b-8641-cf865d76f464"}	2026-07-22 09:05:40.92936
55	allocation	sales_order	\N	3	\N	1	\N	\N	\N	6.000	3	\N	\N	2026-07-22 09:05:40.932442	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "9611e2f7-3042-4d9b-ae57-9410ca6c91c7", "sales_order_id": "97812e45-0a4d-437b-8641-cf865d76f464"}	2026-07-22 09:05:40.932442
56	allocation	sales_order	\N	22	\N	1	\N	\N	\N	7.000	3	\N	\N	2026-07-22 09:05:40.935409	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "d22c4375-10da-4542-a60f-33699b2c47b3", "sales_order_id": "97812e45-0a4d-437b-8641-cf865d76f464"}	2026-07-22 09:05:40.935409
57	allocation	sales_order	\N	22	\N	1	\N	\N	\N	14.000	3	\N	\N	2026-07-22 09:52:19.491342	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "4d51a356-35a5-4d98-a9fc-7649c517971c", "sales_order_id": "d3bb4057-5f3d-4eaa-aa7f-2dd14ecfeebb"}	2026-07-22 09:52:19.491342
58	allocation	sales_order	\N	22	\N	1	\N	\N	\N	14.000	3	\N	\N	2026-07-22 09:54:51.079511	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "59cb0035-152d-438d-9f50-5adcb2cd270b", "sales_order_id": "c2fc4fa8-a714-458d-9656-f82085a9c28b"}	2026-07-22 09:54:51.079511
59	allocation	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-22 09:54:51.083972	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "838e7644-3ce0-4416-b82b-a9b411d2822b", "sales_order_id": "c2fc4fa8-a714-458d-9656-f82085a9c28b"}	2026-07-22 09:54:51.083972
60	sale	sales_order	\N	22	\N	1	\N	\N	\N	14.000	3	\N	\N	2026-07-22 09:58:06.884572	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "59cb0035-152d-438d-9f50-5adcb2cd270b", "sales_order_id": "c2fc4fa8-a714-458d-9656-f82085a9c28b", "sales_order_number": "ORD-20260722095451"}	2026-07-22 09:58:06.884572
61	sale	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-22 09:58:06.884572	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "838e7644-3ce0-4416-b82b-a9b411d2822b", "sales_order_id": "c2fc4fa8-a714-458d-9656-f82085a9c28b", "sales_order_number": "ORD-20260722095451"}	2026-07-22 09:58:06.884572
62	allocation	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-22 10:01:20.560272	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "82c01550-a83f-4810-b103-b97b395c740a", "sales_order_id": "dad48529-23c0-4980-8cfc-648f16494e67"}	2026-07-22 10:01:20.560272
63	allocation	sales_order	\N	26	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-22 10:01:20.565562	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "4558da76-b9ee-4ef4-be21-ce7ea7e03a87", "sales_order_id": "dad48529-23c0-4980-8cfc-648f16494e67"}	2026-07-22 10:01:20.565562
64	allocation	sales_order	\N	6	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-22 10:01:20.56953	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "32a22366-4209-49ed-ad19-e500629785b8", "sales_order_id": "dad48529-23c0-4980-8cfc-648f16494e67"}	2026-07-22 10:01:20.56953
65	allocation	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-22 10:02:26.793413	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "b3860160-4d37-4c85-84ad-7bfa860978ad", "sales_order_id": "98ec425c-3a11-4f1e-ac90-5d96ab6e5819"}	2026-07-22 10:02:26.793413
66	allocation	sales_order	\N	26	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-22 10:02:26.796809	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "28dc3f6b-b9b7-49f5-8e20-8cf60064e7db", "sales_order_id": "98ec425c-3a11-4f1e-ac90-5d96ab6e5819"}	2026-07-22 10:02:26.796809
67	sale	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-22 10:02:38.590571	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "b3860160-4d37-4c85-84ad-7bfa860978ad", "sales_order_id": "98ec425c-3a11-4f1e-ac90-5d96ab6e5819", "sales_order_number": "ORD-20260722100226"}	2026-07-22 10:02:38.590571
68	sale	sales_order	\N	26	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-22 10:02:38.590571	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "28dc3f6b-b9b7-49f5-8e20-8cf60064e7db", "sales_order_id": "98ec425c-3a11-4f1e-ac90-5d96ab6e5819", "sales_order_number": "ORD-20260722100226"}	2026-07-22 10:02:38.590571
69	allocation	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-22 10:02:51.752264	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "ab7fbbb5-b72c-41d9-a10f-88d0657af3ec", "sales_order_id": "b5f91a7e-875e-46c5-94d1-c86525cccd5f"}	2026-07-22 10:02:51.752264
70	sale	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-22 10:02:56.568605	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "ab7fbbb5-b72c-41d9-a10f-88d0657af3ec", "sales_order_id": "b5f91a7e-875e-46c5-94d1-c86525cccd5f", "sales_order_number": "ORD-20260722100251"}	2026-07-22 10:02:56.568605
71	allocation	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-22 10:05:41.927857	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "4b2d61ff-ee32-46aa-b904-83ca73349f46", "sales_order_id": "72138058-5326-4827-a07b-1b92c01446f3"}	2026-07-22 10:05:41.927857
72	allocation	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-22 10:07:14.535207	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "7223c42c-99ac-4ec4-befe-97f074ffb049", "sales_order_id": "489e6dac-d79f-417a-8014-cdc3360080f7"}	2026-07-22 10:07:14.535207
73	allocation	sales_order	\N	40	\N	1	\N	\N	\N	1.000	4	\N	\N	2026-07-22 13:11:05.662225	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "8fdde528-3787-47f9-b092-10ffce38fb4b", "sales_order_id": "11003230-d7b2-44d1-b057-5a5347af254d"}	2026-07-22 13:11:05.662225
74	allocation	sales_order	\N	22	\N	1	\N	\N	\N	6.000	3	\N	\N	2026-07-22 13:11:05.670655	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "72bcc36b-8655-412a-9081-f54ebb3d5bda", "sales_order_id": "11003230-d7b2-44d1-b057-5a5347af254d"}	2026-07-22 13:11:05.670655
75	allocation	sales_order	\N	8	\N	1	\N	\N	\N	10.000	1	\N	\N	2026-07-22 13:11:05.673511	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "9e561912-3f5a-496e-89db-84ebf82f48be", "sales_order_id": "11003230-d7b2-44d1-b057-5a5347af254d"}	2026-07-22 13:11:05.673511
76	sale	sales_order	\N	40	\N	1	\N	\N	\N	1.000	4	\N	\N	2026-07-22 13:11:10.83785	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "8fdde528-3787-47f9-b092-10ffce38fb4b", "sales_order_id": "11003230-d7b2-44d1-b057-5a5347af254d", "sales_order_number": "ORD-20260722131105"}	2026-07-22 13:11:10.83785
77	sale	sales_order	\N	22	\N	1	\N	\N	\N	6.000	3	\N	\N	2026-07-22 13:11:10.83785	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "72bcc36b-8655-412a-9081-f54ebb3d5bda", "sales_order_id": "11003230-d7b2-44d1-b057-5a5347af254d", "sales_order_number": "ORD-20260722131105"}	2026-07-22 13:11:10.83785
78	sale	sales_order	\N	8	\N	1	\N	\N	\N	10.000	1	\N	\N	2026-07-22 13:11:10.83785	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "9e561912-3f5a-496e-89db-84ebf82f48be", "sales_order_id": "11003230-d7b2-44d1-b057-5a5347af254d", "sales_order_number": "ORD-20260722131105"}	2026-07-22 13:11:10.83785
79	allocation	sales_order	\N	8	\N	1	\N	\N	\N	10.000	1	\N	\N	2026-07-22 13:19:08.047146	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "5c44e330-c7be-458f-a091-e75647f46a26", "sales_order_id": "e04934f5-c4d6-4676-9932-8cfa48b76178"}	2026-07-22 13:19:08.047146
80	allocation	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-23 06:18:29.326236	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "df392de0-e4ce-41e7-8734-5d951a583f32", "sales_order_id": "bb618d80-6189-4df6-91a2-263755459002"}	2026-07-23 06:18:29.326236
81	allocation	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-23 06:18:29.334637	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "77eae67e-e5b5-4946-bc10-8012f0df62ae", "sales_order_id": "bb618d80-6189-4df6-91a2-263755459002"}	2026-07-23 06:18:29.334637
82	allocation	sales_order	\N	6	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-23 06:18:29.337417	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "be254388-7573-4d33-a8e9-f0d93dae7b8d", "sales_order_id": "bb618d80-6189-4df6-91a2-263755459002"}	2026-07-23 06:18:29.337417
83	allocation	sales_order	\N	23	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-23 06:18:29.340658	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "5b2ad169-4023-4d2d-ae62-113687769bca", "sales_order_id": "bb618d80-6189-4df6-91a2-263755459002"}	2026-07-23 06:18:29.340658
84	sale	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-23 06:18:34.74545	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "df392de0-e4ce-41e7-8734-5d951a583f32", "sales_order_id": "bb618d80-6189-4df6-91a2-263755459002", "sales_order_number": "ORD-20260723061829"}	2026-07-23 06:18:34.74545
85	sale	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-23 06:18:34.74545	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "77eae67e-e5b5-4946-bc10-8012f0df62ae", "sales_order_id": "bb618d80-6189-4df6-91a2-263755459002", "sales_order_number": "ORD-20260723061829"}	2026-07-23 06:18:34.74545
86	sale	sales_order	\N	6	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-23 06:18:34.74545	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "be254388-7573-4d33-a8e9-f0d93dae7b8d", "sales_order_id": "bb618d80-6189-4df6-91a2-263755459002", "sales_order_number": "ORD-20260723061829"}	2026-07-23 06:18:34.74545
87	sale	sales_order	\N	23	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-23 06:18:34.74545	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "5b2ad169-4023-4d2d-ae62-113687769bca", "sales_order_id": "bb618d80-6189-4df6-91a2-263755459002", "sales_order_number": "ORD-20260723061829"}	2026-07-23 06:18:34.74545
88	allocation	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-23 06:28:34.492641	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "ce9ed16e-ae3b-47c0-b308-0f7cd816052d", "sales_order_id": "fc451cb2-50b8-45d3-80a5-e478c795339a"}	2026-07-23 06:28:34.492641
89	allocation	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-23 06:28:34.497829	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "f93a6f15-bf89-41bc-b4e1-d487e01b45a1", "sales_order_id": "fc451cb2-50b8-45d3-80a5-e478c795339a"}	2026-07-23 06:28:34.497829
90	allocation	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-23 06:28:34.500809	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "69af6da6-6e4a-4395-ab5e-8a8e8064f388", "sales_order_id": "fc451cb2-50b8-45d3-80a5-e478c795339a"}	2026-07-23 06:28:34.500809
91	sale	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-23 06:28:39.288752	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "ce9ed16e-ae3b-47c0-b308-0f7cd816052d", "sales_order_id": "fc451cb2-50b8-45d3-80a5-e478c795339a", "sales_order_number": "ORD-20260723062834"}	2026-07-23 06:28:39.288752
92	sale	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-23 06:28:39.288752	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "f93a6f15-bf89-41bc-b4e1-d487e01b45a1", "sales_order_id": "fc451cb2-50b8-45d3-80a5-e478c795339a", "sales_order_number": "ORD-20260723062834"}	2026-07-23 06:28:39.288752
93	sale	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-23 06:28:39.288752	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "69af6da6-6e4a-4395-ab5e-8a8e8064f388", "sales_order_id": "fc451cb2-50b8-45d3-80a5-e478c795339a", "sales_order_number": "ORD-20260723062834"}	2026-07-23 06:28:39.288752
94	allocation	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-23 06:37:55.631893	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "ac00b2f7-c21e-446f-9bbb-d82930ffb617", "sales_order_id": "604348ff-bd9a-480f-9d9e-381151944683"}	2026-07-23 06:37:55.631893
95	sale	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-23 06:38:00.850442	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "ac00b2f7-c21e-446f-9bbb-d82930ffb617", "sales_order_id": "604348ff-bd9a-480f-9d9e-381151944683", "sales_order_number": "ORD-20260723063755"}	2026-07-23 06:38:00.850442
96	allocation	sales_order	\N	26	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-23 08:31:04.312582	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "6d3bd4d3-e662-49c1-bbf1-8081d61e9efd", "sales_order_id": "1c15059e-0602-4345-8a9e-3d7aec6962a4"}	2026-07-23 08:31:04.312582
97	allocation	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-23 08:31:04.321125	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "00dd119f-0fdf-4be2-b858-6d690b90ed2e", "sales_order_id": "1c15059e-0602-4345-8a9e-3d7aec6962a4"}	2026-07-23 08:31:04.321125
98	allocation	sales_order	\N	27	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-23 08:31:04.324196	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "d93ed5f8-70a5-4ee4-9868-4fe383f5b131", "sales_order_id": "1c15059e-0602-4345-8a9e-3d7aec6962a4"}	2026-07-23 08:31:04.324196
99	sale	sales_order	\N	26	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-23 08:31:13.92435	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "6d3bd4d3-e662-49c1-bbf1-8081d61e9efd", "sales_order_id": "1c15059e-0602-4345-8a9e-3d7aec6962a4", "sales_order_number": "ORD-20260723083104"}	2026-07-23 08:31:13.92435
100	sale	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-23 08:31:13.92435	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "00dd119f-0fdf-4be2-b858-6d690b90ed2e", "sales_order_id": "1c15059e-0602-4345-8a9e-3d7aec6962a4", "sales_order_number": "ORD-20260723083104"}	2026-07-23 08:31:13.92435
101	sale	sales_order	\N	27	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-23 08:31:13.92435	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "d93ed5f8-70a5-4ee4-9868-4fe383f5b131", "sales_order_id": "1c15059e-0602-4345-8a9e-3d7aec6962a4", "sales_order_number": "ORD-20260723083104"}	2026-07-23 08:31:13.92435
102	Transfer	Transfer	202300942	35	\N	1	1	7	6	120.000	10	\N	\N	2026-07-23 09:09:47.212	1	Pending	3.5000	420.00	{"remarks": "hhh", "origin_store": "Qitaf al Ayela", "origin_location": "RYD-BACK - Back Storage", "destination_store": "Qitaf al Ayela", "destination_location": "RYD-HOUSEHOLD - Household Products"}	2026-07-23 09:10:23.208178
103	Transfer	Transfer	202300942	30	\N	4	1	8	5	120.000	2	\N	\N	2026-07-23 09:15:46.256	1	Pending	12.9500	1554.00	{"remarks": "ccc", "origin_store": "Qitaf Warehouse", "origin_location": "WH-ZONE-A - Zone A - Dry Goods", "destination_store": "Qitaf al Ayela", "destination_location": "RYD-PERSONAL - Personal Care"}	2026-07-23 09:16:22.258082
105	Transfer	Transfer	202300942	6	\N	4	1	8	5	160.000	10	\N	\N	2026-07-23 09:15:46.256	1	Pending	12.9500	2072.00	{"remarks": "ccc", "origin_store": "Qitaf Warehouse", "origin_location": "WH-ZONE-A - Zone A - Dry Goods", "destination_store": "Qitaf al Ayela", "destination_location": "RYD-PERSONAL - Personal Care"}	2026-07-23 09:16:22.258713
104	Transfer	Transfer	202300942	14	\N	4	1	8	5	150.000	3	\N	\N	2026-07-23 09:15:46.256	1	Pending	7.9500	1192.50	{"remarks": "ccc", "origin_store": "Qitaf Warehouse", "origin_location": "WH-ZONE-A - Zone A - Dry Goods", "destination_store": "Qitaf al Ayela", "destination_location": "RYD-PERSONAL - Personal Care"}	2026-07-23 09:16:22.258842
106	allocation	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-23 10:54:04.972849	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "02726bf2-c339-485c-a36b-d2490f4b03fb", "sales_order_id": "275b4bb0-20a8-430e-be00-7e0ee053f49f"}	2026-07-23 10:54:04.972849
107	sale	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-23 10:54:37.846937	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "02726bf2-c339-485c-a36b-d2490f4b03fb", "sales_order_id": "275b4bb0-20a8-430e-be00-7e0ee053f49f", "sales_order_number": "ORD-20260723105404"}	2026-07-23 10:54:37.846937
108	allocation	sales_order	\N	22	\N	1	\N	\N	\N	6.000	3	\N	\N	2026-07-23 11:07:35.563214	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "75c669c9-4150-47b6-b976-23845f0fa70b", "sales_order_id": "25f422ca-343b-4dac-b452-96f9e58dbc82"}	2026-07-23 11:07:35.563214
109	allocation	sales_order	\N	22	\N	1	\N	\N	\N	6.000	3	\N	\N	2026-07-23 11:07:35.568491	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "29a02373-9c9f-4689-856f-fbfe4941396e", "sales_order_id": "25f422ca-343b-4dac-b452-96f9e58dbc82"}	2026-07-23 11:07:35.568491
110	allocation	sales_order	\N	40	\N	1	\N	\N	\N	6.000	4	\N	\N	2026-07-23 11:07:35.571113	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "b369d142-c1a9-4d62-b525-8cf3da80901a", "sales_order_id": "25f422ca-343b-4dac-b452-96f9e58dbc82"}	2026-07-23 11:07:35.571113
111	sale	sales_order	\N	22	\N	1	\N	\N	\N	6.000	3	\N	\N	2026-07-23 11:07:40.233784	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "75c669c9-4150-47b6-b976-23845f0fa70b", "sales_order_id": "25f422ca-343b-4dac-b452-96f9e58dbc82", "sales_order_number": "ORD-20260723110735"}	2026-07-23 11:07:40.233784
112	sale	sales_order	\N	22	\N	1	\N	\N	\N	6.000	3	\N	\N	2026-07-23 11:07:40.233784	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "29a02373-9c9f-4689-856f-fbfe4941396e", "sales_order_id": "25f422ca-343b-4dac-b452-96f9e58dbc82", "sales_order_number": "ORD-20260723110735"}	2026-07-23 11:07:40.233784
113	sale	sales_order	\N	40	\N	1	\N	\N	\N	6.000	4	\N	\N	2026-07-23 11:07:40.233784	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "b369d142-c1a9-4d62-b525-8cf3da80901a", "sales_order_id": "25f422ca-343b-4dac-b452-96f9e58dbc82", "sales_order_number": "ORD-20260723110735"}	2026-07-23 11:07:40.233784
114	allocation	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-23 11:16:48.147847	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "61c3b086-fa50-4878-807d-d23a3f76a42f", "sales_order_id": "60f69ca5-9235-44c5-bfd6-acb162e8e2c3"}	2026-07-23 11:16:48.147847
115	allocation	sales_order	\N	26	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-23 11:16:48.157477	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "1dc7592a-291f-4112-9a44-f9484a9a0542", "sales_order_id": "60f69ca5-9235-44c5-bfd6-acb162e8e2c3"}	2026-07-23 11:16:48.157477
116	allocation	sales_order	\N	40	\N	1	\N	\N	\N	1.000	4	\N	\N	2026-07-23 11:16:48.160582	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "1676583f-3457-4367-bccb-e3bc1d7cfe17", "sales_order_id": "60f69ca5-9235-44c5-bfd6-acb162e8e2c3"}	2026-07-23 11:16:48.160582
117	allocation	sales_order	\N	40	\N	1	\N	\N	\N	3.000	4	\N	\N	2026-07-23 11:16:48.163745	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "b37b9784-5df4-4af3-9599-4b77e717baf6", "sales_order_id": "60f69ca5-9235-44c5-bfd6-acb162e8e2c3"}	2026-07-23 11:16:48.163745
118	sale	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-23 11:17:01.368133	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "61c3b086-fa50-4878-807d-d23a3f76a42f", "sales_order_id": "60f69ca5-9235-44c5-bfd6-acb162e8e2c3", "sales_order_number": "ORD-20260723111648"}	2026-07-23 11:17:01.368133
119	sale	sales_order	\N	26	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-23 11:17:01.368133	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "1dc7592a-291f-4112-9a44-f9484a9a0542", "sales_order_id": "60f69ca5-9235-44c5-bfd6-acb162e8e2c3", "sales_order_number": "ORD-20260723111648"}	2026-07-23 11:17:01.368133
120	sale	sales_order	\N	40	\N	1	\N	\N	\N	1.000	4	\N	\N	2026-07-23 11:17:01.368133	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "1676583f-3457-4367-bccb-e3bc1d7cfe17", "sales_order_id": "60f69ca5-9235-44c5-bfd6-acb162e8e2c3", "sales_order_number": "ORD-20260723111648"}	2026-07-23 11:17:01.368133
121	sale	sales_order	\N	40	\N	1	\N	\N	\N	3.000	4	\N	\N	2026-07-23 11:17:01.368133	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "b37b9784-5df4-4af3-9599-4b77e717baf6", "sales_order_id": "60f69ca5-9235-44c5-bfd6-acb162e8e2c3", "sales_order_number": "ORD-20260723111648"}	2026-07-23 11:17:01.368133
122	allocation	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-23 11:24:59.689522	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "2d16c051-15c8-4711-9907-6f0aa13598e1", "sales_order_id": "28789b9c-918d-4697-ae9c-19feed267c18"}	2026-07-23 11:24:59.689522
123	allocation	sales_order	\N	15	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-23 11:24:59.693941	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "b4704a02-23db-49c9-892b-f3a60828bc92", "sales_order_id": "28789b9c-918d-4697-ae9c-19feed267c18"}	2026-07-23 11:24:59.693941
124	sale	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-23 11:25:10.385199	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "2d16c051-15c8-4711-9907-6f0aa13598e1", "sales_order_id": "28789b9c-918d-4697-ae9c-19feed267c18", "sales_order_number": "ORD-20260723112459"}	2026-07-23 11:25:10.385199
125	sale	sales_order	\N	15	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-23 11:25:10.385199	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "b4704a02-23db-49c9-892b-f3a60828bc92", "sales_order_id": "28789b9c-918d-4697-ae9c-19feed267c18", "sales_order_number": "ORD-20260723112459"}	2026-07-23 11:25:10.385199
126	allocation	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-23 11:42:46.155412	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "4728f027-3c9d-455c-b055-1a65ab75b47f", "sales_order_id": "75b6b802-225e-4481-8ef4-cabb0bbfd07c"}	2026-07-23 11:42:46.155412
127	sale	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-23 11:42:50.044955	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "4728f027-3c9d-455c-b055-1a65ab75b47f", "sales_order_id": "75b6b802-225e-4481-8ef4-cabb0bbfd07c", "sales_order_number": "ORD-20260723114246"}	2026-07-23 11:42:50.044955
128	allocation	sales_order	\N	33	\N	1	\N	\N	\N	2.000	10	\N	\N	2026-07-24 09:14:58.917584	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "d16fa332-340e-43a2-807b-f1e8ec8c0d61", "sales_order_id": "2b000f45-99b0-4a8a-b4c7-60dcf39af0b6"}	2026-07-24 09:14:58.917584
129	allocation	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-24 09:14:58.927071	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "58a71e00-b707-4ed1-aa7f-b7b91c9919d5", "sales_order_id": "2b000f45-99b0-4a8a-b4c7-60dcf39af0b6"}	2026-07-24 09:14:58.927071
130	allocation	sales_order	\N	27	\N	1	\N	\N	\N	2.000	8	\N	\N	2026-07-24 09:14:58.930078	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "8816b1b7-3e5f-4f8e-a0f8-0c03fcb55a04", "sales_order_id": "2b000f45-99b0-4a8a-b4c7-60dcf39af0b6"}	2026-07-24 09:14:58.930078
131	allocation	sales_order	\N	6	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-24 09:14:58.933355	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "5dd917b2-2415-4af4-a0f2-07e3fe6b916a", "sales_order_id": "2b000f45-99b0-4a8a-b4c7-60dcf39af0b6"}	2026-07-24 09:14:58.933355
132	allocation	sales_order	\N	23	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-24 09:14:58.936629	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "9f9e1bda-4264-4bb6-882f-5188177710b6", "sales_order_id": "2b000f45-99b0-4a8a-b4c7-60dcf39af0b6"}	2026-07-24 09:14:58.936629
133	sale	sales_order	\N	33	\N	1	\N	\N	\N	2.000	10	\N	\N	2026-07-24 09:15:04.033643	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "d16fa332-340e-43a2-807b-f1e8ec8c0d61", "sales_order_id": "2b000f45-99b0-4a8a-b4c7-60dcf39af0b6", "sales_order_number": "ORD-20260724091458"}	2026-07-24 09:15:04.033643
134	sale	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-24 09:15:04.033643	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "58a71e00-b707-4ed1-aa7f-b7b91c9919d5", "sales_order_id": "2b000f45-99b0-4a8a-b4c7-60dcf39af0b6", "sales_order_number": "ORD-20260724091458"}	2026-07-24 09:15:04.033643
135	sale	sales_order	\N	27	\N	1	\N	\N	\N	2.000	8	\N	\N	2026-07-24 09:15:04.033643	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "8816b1b7-3e5f-4f8e-a0f8-0c03fcb55a04", "sales_order_id": "2b000f45-99b0-4a8a-b4c7-60dcf39af0b6", "sales_order_number": "ORD-20260724091458"}	2026-07-24 09:15:04.033643
136	sale	sales_order	\N	6	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-24 09:15:04.033643	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "5dd917b2-2415-4af4-a0f2-07e3fe6b916a", "sales_order_id": "2b000f45-99b0-4a8a-b4c7-60dcf39af0b6", "sales_order_number": "ORD-20260724091458"}	2026-07-24 09:15:04.033643
137	sale	sales_order	\N	23	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-24 09:15:04.033643	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "9f9e1bda-4264-4bb6-882f-5188177710b6", "sales_order_id": "2b000f45-99b0-4a8a-b4c7-60dcf39af0b6", "sales_order_number": "ORD-20260724091458"}	2026-07-24 09:15:04.033643
138	allocation	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-27 10:47:45.076747	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "86efd7c0-43bd-4be8-b080-acb80a55b0ee", "sales_order_id": "17ca1d02-1e68-4055-8370-93d99ef0a469"}	2026-07-27 10:47:45.076747
139	sale	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-27 10:48:00.830681	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "86efd7c0-43bd-4be8-b080-acb80a55b0ee", "sales_order_id": "17ca1d02-1e68-4055-8370-93d99ef0a469", "sales_order_number": "ORD-20260727104745"}	2026-07-27 10:48:00.830681
140	allocation	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-27 11:06:26.245697	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "059badd2-9457-42b8-836a-153f9a685b5a", "sales_order_id": "2bd4df2f-49f8-409f-aaaa-559039b645fd"}	2026-07-27 11:06:26.245697
141	allocation	sales_order	\N	26	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-27 11:06:26.256226	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "b828cf32-51d1-4542-8890-6f5f626c5fac", "sales_order_id": "2bd4df2f-49f8-409f-aaaa-559039b645fd"}	2026-07-27 11:06:26.256226
142	sale	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-27 11:06:39.759726	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "059badd2-9457-42b8-836a-153f9a685b5a", "sales_order_id": "2bd4df2f-49f8-409f-aaaa-559039b645fd", "sales_order_number": "ORD-20260727110626"}	2026-07-27 11:06:39.759726
143	sale	sales_order	\N	26	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-27 11:06:39.759726	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "b828cf32-51d1-4542-8890-6f5f626c5fac", "sales_order_id": "2bd4df2f-49f8-409f-aaaa-559039b645fd", "sales_order_number": "ORD-20260727110626"}	2026-07-27 11:06:39.759726
144	allocation	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-27 11:11:19.929954	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "8aa59ec5-10c7-455f-9351-acd48c8b274a", "sales_order_id": "d11c1e2b-df97-465b-bff1-34b6780ff4fc"}	2026-07-27 11:11:19.929954
145	sale	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-27 11:11:26.155059	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "8aa59ec5-10c7-455f-9351-acd48c8b274a", "sales_order_id": "d11c1e2b-df97-465b-bff1-34b6780ff4fc", "sales_order_number": "ORD-20260727111119"}	2026-07-27 11:11:26.155059
146	allocation	sales_order	\N	17	\N	1	\N	\N	\N	1.000	4	\N	\N	2026-07-27 11:42:20.354351	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "bac5e6cc-1f4a-4192-a4b8-95df92155810", "sales_order_id": "358a5da2-fb5c-496d-b582-a1fd790b4535"}	2026-07-27 11:42:20.354351
147	allocation	sales_order	\N	29	\N	1	\N	\N	\N	1.000	2	\N	\N	2026-07-27 11:42:20.362929	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "e9a032e2-63fa-4afc-8145-374bd974c44b", "sales_order_id": "358a5da2-fb5c-496d-b582-a1fd790b4535"}	2026-07-27 11:42:20.362929
148	allocation	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-27 11:42:20.366366	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "b5a1cb2c-e6d4-45f3-942a-92782ca6f820", "sales_order_id": "358a5da2-fb5c-496d-b582-a1fd790b4535"}	2026-07-27 11:42:20.366366
149	allocation	sales_order	\N	24	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-27 11:42:20.369583	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "42dcf4a1-4b31-4c0e-a002-b60b541ffa91", "sales_order_id": "358a5da2-fb5c-496d-b582-a1fd790b4535"}	2026-07-27 11:42:20.369583
150	allocation	sales_order	\N	35	\N	1	\N	\N	\N	8.000	10	\N	\N	2026-07-27 11:42:20.373045	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "72bccde1-fc48-4ff4-ad34-f0f271d2f9c7", "sales_order_id": "358a5da2-fb5c-496d-b582-a1fd790b4535"}	2026-07-27 11:42:20.373045
151	sale	sales_order	\N	17	\N	1	\N	\N	\N	1.000	4	\N	\N	2026-07-27 11:42:29.409901	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "bac5e6cc-1f4a-4192-a4b8-95df92155810", "sales_order_id": "358a5da2-fb5c-496d-b582-a1fd790b4535", "sales_order_number": "ORD-20260727114220"}	2026-07-27 11:42:29.409901
152	sale	sales_order	\N	29	\N	1	\N	\N	\N	1.000	2	\N	\N	2026-07-27 11:42:29.409901	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "e9a032e2-63fa-4afc-8145-374bd974c44b", "sales_order_id": "358a5da2-fb5c-496d-b582-a1fd790b4535", "sales_order_number": "ORD-20260727114220"}	2026-07-27 11:42:29.409901
153	sale	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-27 11:42:29.409901	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "b5a1cb2c-e6d4-45f3-942a-92782ca6f820", "sales_order_id": "358a5da2-fb5c-496d-b582-a1fd790b4535", "sales_order_number": "ORD-20260727114220"}	2026-07-27 11:42:29.409901
154	sale	sales_order	\N	24	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-27 11:42:29.409901	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "42dcf4a1-4b31-4c0e-a002-b60b541ffa91", "sales_order_id": "358a5da2-fb5c-496d-b582-a1fd790b4535", "sales_order_number": "ORD-20260727114220"}	2026-07-27 11:42:29.409901
155	sale	sales_order	\N	35	\N	1	\N	\N	\N	8.000	10	\N	\N	2026-07-27 11:42:29.409901	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "72bccde1-fc48-4ff4-ad34-f0f271d2f9c7", "sales_order_id": "358a5da2-fb5c-496d-b582-a1fd790b4535", "sales_order_number": "ORD-20260727114220"}	2026-07-27 11:42:29.409901
156	allocation	sales_order	\N	14	\N	1	\N	\N	\N	6.000	3	\N	\N	2026-07-27 11:56:46.14556	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "60731753-2b8d-4d19-a499-20f7f8bda46e", "sales_order_id": "4aa69799-b182-41e0-9fba-92c284472094"}	2026-07-27 11:56:46.14556
157	allocation	sales_order	\N	33	\N	1	\N	\N	\N	43.000	10	\N	\N	2026-07-27 11:56:46.15391	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "86496162-549c-41d4-b6cb-0ac492c0c685", "sales_order_id": "4aa69799-b182-41e0-9fba-92c284472094"}	2026-07-27 11:56:46.15391
158	allocation	sales_order	\N	24	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-27 11:56:46.157259	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "a5e9434a-63dd-4d5c-bcae-959cf5ba5164", "sales_order_id": "4aa69799-b182-41e0-9fba-92c284472094"}	2026-07-27 11:56:46.157259
159	allocation	sales_order	\N	24	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-27 11:56:46.16053	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "13aa31f8-d690-411d-94e0-8ed9d507c16e", "sales_order_id": "4aa69799-b182-41e0-9fba-92c284472094"}	2026-07-27 11:56:46.16053
160	allocation	sales_order	\N	24	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-27 11:56:46.163758	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "f67436b0-8c7c-471d-a716-b6a17f896f73", "sales_order_id": "4aa69799-b182-41e0-9fba-92c284472094"}	2026-07-27 11:56:46.163758
161	allocation	sales_order	\N	29	\N	1	\N	\N	\N	1.000	2	\N	\N	2026-07-27 11:56:46.16723	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "c9846bef-de9c-4992-93f1-dd3dccbd820d", "sales_order_id": "4aa69799-b182-41e0-9fba-92c284472094"}	2026-07-27 11:56:46.16723
162	allocation	sales_order	\N	29	\N	1	\N	\N	\N	1.000	2	\N	\N	2026-07-27 11:56:46.170604	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "f213ea8e-22f7-4072-9013-73a564fe1866", "sales_order_id": "4aa69799-b182-41e0-9fba-92c284472094"}	2026-07-27 11:56:46.170604
163	allocation	sales_order	\N	29	\N	1	\N	\N	\N	1.000	2	\N	\N	2026-07-27 11:56:46.173041	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "5e1fca8e-e342-4e7b-b88c-1a452af016ec", "sales_order_id": "4aa69799-b182-41e0-9fba-92c284472094"}	2026-07-27 11:56:46.173041
164	sale	sales_order	\N	14	\N	1	\N	\N	\N	6.000	3	\N	\N	2026-07-27 11:56:51.0094	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "60731753-2b8d-4d19-a499-20f7f8bda46e", "sales_order_id": "4aa69799-b182-41e0-9fba-92c284472094", "sales_order_number": "ORD-20260727115646"}	2026-07-27 11:56:51.0094
165	sale	sales_order	\N	33	\N	1	\N	\N	\N	43.000	10	\N	\N	2026-07-27 11:56:51.0094	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "86496162-549c-41d4-b6cb-0ac492c0c685", "sales_order_id": "4aa69799-b182-41e0-9fba-92c284472094", "sales_order_number": "ORD-20260727115646"}	2026-07-27 11:56:51.0094
166	sale	sales_order	\N	24	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-27 11:56:51.0094	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "a5e9434a-63dd-4d5c-bcae-959cf5ba5164", "sales_order_id": "4aa69799-b182-41e0-9fba-92c284472094", "sales_order_number": "ORD-20260727115646"}	2026-07-27 11:56:51.0094
167	sale	sales_order	\N	24	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-27 11:56:51.0094	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "13aa31f8-d690-411d-94e0-8ed9d507c16e", "sales_order_id": "4aa69799-b182-41e0-9fba-92c284472094", "sales_order_number": "ORD-20260727115646"}	2026-07-27 11:56:51.0094
168	sale	sales_order	\N	24	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-27 11:56:51.0094	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "f67436b0-8c7c-471d-a716-b6a17f896f73", "sales_order_id": "4aa69799-b182-41e0-9fba-92c284472094", "sales_order_number": "ORD-20260727115646"}	2026-07-27 11:56:51.0094
169	sale	sales_order	\N	29	\N	1	\N	\N	\N	1.000	2	\N	\N	2026-07-27 11:56:51.0094	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "c9846bef-de9c-4992-93f1-dd3dccbd820d", "sales_order_id": "4aa69799-b182-41e0-9fba-92c284472094", "sales_order_number": "ORD-20260727115646"}	2026-07-27 11:56:51.0094
170	sale	sales_order	\N	29	\N	1	\N	\N	\N	1.000	2	\N	\N	2026-07-27 11:56:51.0094	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "f213ea8e-22f7-4072-9013-73a564fe1866", "sales_order_id": "4aa69799-b182-41e0-9fba-92c284472094", "sales_order_number": "ORD-20260727115646"}	2026-07-27 11:56:51.0094
171	sale	sales_order	\N	29	\N	1	\N	\N	\N	1.000	2	\N	\N	2026-07-27 11:56:51.0094	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "5e1fca8e-e342-4e7b-b88c-1a452af016ec", "sales_order_id": "4aa69799-b182-41e0-9fba-92c284472094", "sales_order_number": "ORD-20260727115646"}	2026-07-27 11:56:51.0094
172	allocation	sales_order	\N	29	\N	1	\N	\N	\N	8.000	2	\N	\N	2026-07-27 12:05:02.739341	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "a60f6c1a-7870-461a-afff-9fdebaf6f01a", "sales_order_id": "f556b2b5-b3c8-4f70-8ccb-cc09f953a861"}	2026-07-27 12:05:02.739341
173	allocation	sales_order	\N	35	\N	1	\N	\N	\N	39.000	10	\N	\N	2026-07-27 12:05:02.74342	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "22de0ac7-6f07-4723-914a-30adeeabf3d5", "sales_order_id": "f556b2b5-b3c8-4f70-8ccb-cc09f953a861"}	2026-07-27 12:05:02.74342
174	allocation	sales_order	\N	14	\N	1	\N	\N	\N	55.000	3	\N	\N	2026-07-27 12:05:02.745548	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "85701d41-e0e3-4619-8a20-27f0107dc66a", "sales_order_id": "f556b2b5-b3c8-4f70-8ccb-cc09f953a861"}	2026-07-27 12:05:02.745548
175	allocation	sales_order	\N	34	\N	1	\N	\N	\N	2.000	11	\N	\N	2026-07-27 12:05:02.747893	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "fca0583b-24cf-4e50-9295-f00d1e670ef5", "sales_order_id": "f556b2b5-b3c8-4f70-8ccb-cc09f953a861"}	2026-07-27 12:05:02.747893
176	allocation	sales_order	\N	17	\N	1	\N	\N	\N	3.000	4	\N	\N	2026-07-27 12:05:02.750493	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "1de243a7-84bb-45db-9d9f-a3a307c3e918", "sales_order_id": "f556b2b5-b3c8-4f70-8ccb-cc09f953a861"}	2026-07-27 12:05:02.750493
177	allocation	sales_order	\N	33	\N	1	\N	\N	\N	7.000	10	\N	\N	2026-07-27 12:05:02.752932	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "17779df2-86ed-42a7-a2af-0f70d9b03b2c", "sales_order_id": "f556b2b5-b3c8-4f70-8ccb-cc09f953a861"}	2026-07-27 12:05:02.752932
178	sale	sales_order	\N	29	\N	1	\N	\N	\N	8.000	2	\N	\N	2026-07-27 12:05:08.876724	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "a60f6c1a-7870-461a-afff-9fdebaf6f01a", "sales_order_id": "f556b2b5-b3c8-4f70-8ccb-cc09f953a861", "sales_order_number": "ORD-20260727120502"}	2026-07-27 12:05:08.876724
179	sale	sales_order	\N	35	\N	1	\N	\N	\N	39.000	10	\N	\N	2026-07-27 12:05:08.876724	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "22de0ac7-6f07-4723-914a-30adeeabf3d5", "sales_order_id": "f556b2b5-b3c8-4f70-8ccb-cc09f953a861", "sales_order_number": "ORD-20260727120502"}	2026-07-27 12:05:08.876724
180	sale	sales_order	\N	14	\N	1	\N	\N	\N	55.000	3	\N	\N	2026-07-27 12:05:08.876724	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "85701d41-e0e3-4619-8a20-27f0107dc66a", "sales_order_id": "f556b2b5-b3c8-4f70-8ccb-cc09f953a861", "sales_order_number": "ORD-20260727120502"}	2026-07-27 12:05:08.876724
181	sale	sales_order	\N	34	\N	1	\N	\N	\N	2.000	11	\N	\N	2026-07-27 12:05:08.876724	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "fca0583b-24cf-4e50-9295-f00d1e670ef5", "sales_order_id": "f556b2b5-b3c8-4f70-8ccb-cc09f953a861", "sales_order_number": "ORD-20260727120502"}	2026-07-27 12:05:08.876724
182	sale	sales_order	\N	17	\N	1	\N	\N	\N	3.000	4	\N	\N	2026-07-27 12:05:08.876724	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "1de243a7-84bb-45db-9d9f-a3a307c3e918", "sales_order_id": "f556b2b5-b3c8-4f70-8ccb-cc09f953a861", "sales_order_number": "ORD-20260727120502"}	2026-07-27 12:05:08.876724
183	sale	sales_order	\N	33	\N	1	\N	\N	\N	7.000	10	\N	\N	2026-07-27 12:05:08.876724	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "17779df2-86ed-42a7-a2af-0f70d9b03b2c", "sales_order_id": "f556b2b5-b3c8-4f70-8ccb-cc09f953a861", "sales_order_number": "ORD-20260727120502"}	2026-07-27 12:05:08.876724
184	allocation	sales_order	\N	14	\N	1	\N	\N	\N	5.000	3	\N	\N	2026-07-27 12:06:45.381384	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "75e5e551-83db-44f8-812e-8223afadca4d", "sales_order_id": "e983e5b0-8b05-4ace-95b8-b629a31013ac"}	2026-07-27 12:06:45.381384
185	allocation	sales_order	\N	35	\N	1	\N	\N	\N	8.000	10	\N	\N	2026-07-27 12:06:45.391426	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "77a28b98-e831-44d0-9760-0242602d0a09", "sales_order_id": "e983e5b0-8b05-4ace-95b8-b629a31013ac"}	2026-07-27 12:06:45.391426
186	allocation	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-27 12:06:45.394773	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "caf8a22d-bd60-4c0b-8619-f20ecbf5cdbb", "sales_order_id": "e983e5b0-8b05-4ace-95b8-b629a31013ac"}	2026-07-27 12:06:45.394773
187	allocation	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-27 12:06:45.397895	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "2de0ea9b-dc60-4508-8be6-15cfa27f3616", "sales_order_id": "e983e5b0-8b05-4ace-95b8-b629a31013ac"}	2026-07-27 12:06:45.397895
188	allocation	sales_order	\N	26	\N	1	\N	\N	\N	5.000	8	\N	\N	2026-07-27 12:06:45.401502	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "7679b75a-e07f-4561-bc75-66b29c4712ff", "sales_order_id": "e983e5b0-8b05-4ace-95b8-b629a31013ac"}	2026-07-27 12:06:45.401502
189	sale	sales_order	\N	14	\N	1	\N	\N	\N	5.000	3	\N	\N	2026-07-27 12:06:50.029317	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "75e5e551-83db-44f8-812e-8223afadca4d", "sales_order_id": "e983e5b0-8b05-4ace-95b8-b629a31013ac", "sales_order_number": "ORD-20260727120645"}	2026-07-27 12:06:50.029317
190	sale	sales_order	\N	35	\N	1	\N	\N	\N	8.000	10	\N	\N	2026-07-27 12:06:50.029317	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "77a28b98-e831-44d0-9760-0242602d0a09", "sales_order_id": "e983e5b0-8b05-4ace-95b8-b629a31013ac", "sales_order_number": "ORD-20260727120645"}	2026-07-27 12:06:50.029317
191	sale	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-27 12:06:50.029317	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "caf8a22d-bd60-4c0b-8619-f20ecbf5cdbb", "sales_order_id": "e983e5b0-8b05-4ace-95b8-b629a31013ac", "sales_order_number": "ORD-20260727120645"}	2026-07-27 12:06:50.029317
192	sale	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-27 12:06:50.029317	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "2de0ea9b-dc60-4508-8be6-15cfa27f3616", "sales_order_id": "e983e5b0-8b05-4ace-95b8-b629a31013ac", "sales_order_number": "ORD-20260727120645"}	2026-07-27 12:06:50.029317
193	sale	sales_order	\N	26	\N	1	\N	\N	\N	5.000	8	\N	\N	2026-07-27 12:06:50.029317	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "7679b75a-e07f-4561-bc75-66b29c4712ff", "sales_order_id": "e983e5b0-8b05-4ace-95b8-b629a31013ac", "sales_order_number": "ORD-20260727120645"}	2026-07-27 12:06:50.029317
194	allocation	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-28 08:35:54.508952	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "f8ee3da4-8c9a-476f-b81d-b1e297e603f6", "sales_order_id": "6fd4ff9a-45d4-4ed7-8002-8b58edc2c9ae"}	2026-07-28 08:35:54.508952
195	allocation	sales_order	\N	27	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-28 08:35:54.517455	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "a4f1874c-5779-4939-8c63-20f190ac1e50", "sales_order_id": "6fd4ff9a-45d4-4ed7-8002-8b58edc2c9ae"}	2026-07-28 08:35:54.517455
196	allocation	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-28 08:35:54.520271	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "d7bb72eb-e3ba-4858-9d7b-36e296e0f07d", "sales_order_id": "6fd4ff9a-45d4-4ed7-8002-8b58edc2c9ae"}	2026-07-28 08:35:54.520271
197	allocation	sales_order	\N	6	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-28 08:35:54.52326	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "13e588a2-79e8-4018-8f23-87716fc002d2", "sales_order_id": "6fd4ff9a-45d4-4ed7-8002-8b58edc2c9ae"}	2026-07-28 08:35:54.52326
198	sale	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-28 08:35:58.826576	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "f8ee3da4-8c9a-476f-b81d-b1e297e603f6", "sales_order_id": "6fd4ff9a-45d4-4ed7-8002-8b58edc2c9ae", "sales_order_number": "ORD-20260728083554"}	2026-07-28 08:35:58.826576
199	sale	sales_order	\N	27	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-28 08:35:58.826576	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "a4f1874c-5779-4939-8c63-20f190ac1e50", "sales_order_id": "6fd4ff9a-45d4-4ed7-8002-8b58edc2c9ae", "sales_order_number": "ORD-20260728083554"}	2026-07-28 08:35:58.826576
200	sale	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-28 08:35:58.826576	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "d7bb72eb-e3ba-4858-9d7b-36e296e0f07d", "sales_order_id": "6fd4ff9a-45d4-4ed7-8002-8b58edc2c9ae", "sales_order_number": "ORD-20260728083554"}	2026-07-28 08:35:58.826576
201	sale	sales_order	\N	6	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-28 08:35:58.826576	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "13e588a2-79e8-4018-8f23-87716fc002d2", "sales_order_id": "6fd4ff9a-45d4-4ed7-8002-8b58edc2c9ae", "sales_order_number": "ORD-20260728083554"}	2026-07-28 08:35:58.826576
202	allocation	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-28 08:37:14.658897	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "a1b75058-9cfe-4d76-8ce2-d0a750133296", "sales_order_id": "2289e488-034e-4715-803e-93febd00c31b"}	2026-07-28 08:37:14.658897
203	allocation	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-28 08:37:14.666542	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "367d9852-ae7b-4d5e-ba7c-7214cb3f07c4", "sales_order_id": "2289e488-034e-4715-803e-93febd00c31b"}	2026-07-28 08:37:14.666542
204	allocation	sales_order	\N	7	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-28 08:37:14.669624	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "5916eedf-ae98-4bce-ac92-5de5d7426987", "sales_order_id": "2289e488-034e-4715-803e-93febd00c31b"}	2026-07-28 08:37:14.669624
205	allocation	sales_order	\N	22	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-28 08:37:14.672764	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "49210164-736e-400f-b21e-f4a9690b59de", "sales_order_id": "2289e488-034e-4715-803e-93febd00c31b"}	2026-07-28 08:37:14.672764
206	sale	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-28 08:37:18.478419	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "a1b75058-9cfe-4d76-8ce2-d0a750133296", "sales_order_id": "2289e488-034e-4715-803e-93febd00c31b", "sales_order_number": "ORD-20260728083714"}	2026-07-28 08:37:18.478419
207	sale	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-28 08:37:18.478419	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "367d9852-ae7b-4d5e-ba7c-7214cb3f07c4", "sales_order_id": "2289e488-034e-4715-803e-93febd00c31b", "sales_order_number": "ORD-20260728083714"}	2026-07-28 08:37:18.478419
208	sale	sales_order	\N	7	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-28 08:37:18.478419	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "5916eedf-ae98-4bce-ac92-5de5d7426987", "sales_order_id": "2289e488-034e-4715-803e-93febd00c31b", "sales_order_number": "ORD-20260728083714"}	2026-07-28 08:37:18.478419
209	sale	sales_order	\N	22	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-28 08:37:18.478419	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "49210164-736e-400f-b21e-f4a9690b59de", "sales_order_id": "2289e488-034e-4715-803e-93febd00c31b", "sales_order_number": "ORD-20260728083714"}	2026-07-28 08:37:18.478419
210	transfer_out	transfer_request	6	1	\N	1	2	\N	\N	5.000	\N	\N	\N	2026-07-28 13:16:15.625563	1	shipped	\N	\N	{"transfer_number": "TR-1003"}	2026-07-28 13:16:15.625563
211	transfer_in	transfer_request	6	1	\N	1	2	\N	\N	5.000	\N	\N	\N	2026-07-28 13:17:43.081608	2	completed	\N	\N	{"transfer_number": "TR-1003"}	2026-07-28 13:17:43.081608
212	transfer_out	transfer_request	8	31	\N	4	1	8	6	10.000	10	\N	\N	2026-07-30 07:30:33.454023	1	shipped	\N	\N	{"transfer_number": "TRF-AEDA-2208"}	2026-07-30 07:30:33.454023
213	transfer_out	transfer_request	8	30	\N	4	1	8	6	20.000	2	\N	\N	2026-07-30 07:30:33.454023	1	shipped	\N	\N	{"transfer_number": "TRF-AEDA-2208"}	2026-07-30 07:30:33.454023
214	transfer_in	transfer_request	8	31	\N	4	1	8	6	10.000	10	\N	\N	2026-07-30 07:47:28.96031	1	completed	\N	\N	{"transfer_number": "TRF-AEDA-2208"}	2026-07-30 07:47:28.96031
215	transfer_in	transfer_request	8	30	\N	4	1	8	6	20.000	2	\N	\N	2026-07-30 07:47:28.96031	1	completed	\N	\N	{"transfer_number": "TRF-AEDA-2208"}	2026-07-30 07:47:28.96031
216	transfer_out	transfer_request	5	23	\N	4	1	8	6	11.000	3	\N	\N	2026-07-30 07:47:50.151711	1	shipped	\N	\N	{"transfer_number": "TRF-2023-00947"}	2026-07-30 07:47:50.151711
217	allocation	sales_order	\N	22	\N	10	\N	\N	\N	6.000	3	\N	\N	2026-07-30 08:02:53.780848	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "a57d2997-0e31-444c-86cb-4c9d311db835", "sales_order_id": "27b796e5-3c0e-40f5-810e-027abd4a7ae4"}	2026-07-30 08:02:53.780848
218	sale	sales_order	\N	22	\N	10	\N	\N	\N	6.000	3	\N	\N	2026-07-30 08:02:59.314065	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "a57d2997-0e31-444c-86cb-4c9d311db835", "sales_order_id": "27b796e5-3c0e-40f5-810e-027abd4a7ae4", "sales_order_number": "ORD-20260730080253"}	2026-07-30 08:02:59.314065
219	allocation	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-30 08:10:35.377004	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "eb5fe8d7-8e72-490d-92f3-a0f3fdd7dbdd", "sales_order_id": "5fd9d2c3-7306-4627-994e-7fbfd52e50a0"}	2026-07-30 08:10:35.377004
220	allocation	sales_order	\N	26	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-30 08:10:35.381567	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "eaf38462-483d-4168-ba61-44bdecf2d407", "sales_order_id": "5fd9d2c3-7306-4627-994e-7fbfd52e50a0"}	2026-07-30 08:10:35.381567
221	sale	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-30 08:19:56.098533	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "eb5fe8d7-8e72-490d-92f3-a0f3fdd7dbdd", "sales_order_id": "5fd9d2c3-7306-4627-994e-7fbfd52e50a0", "sales_order_number": "ORD-20260730081035"}	2026-07-30 08:19:56.098533
222	sale	sales_order	\N	26	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-30 08:19:56.098533	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "eaf38462-483d-4168-ba61-44bdecf2d407", "sales_order_id": "5fd9d2c3-7306-4627-994e-7fbfd52e50a0", "sales_order_number": "ORD-20260730081035"}	2026-07-30 08:19:56.098533
223	allocation	sales_order	\N	34	\N	1	\N	\N	\N	3.000	11	\N	\N	2026-07-30 10:59:28.619179	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "b979d04d-d17a-41e2-b79c-0f52ac94db89", "sales_order_id": "5a6f336d-d29c-45b1-955e-e19af7441324"}	2026-07-30 10:59:28.619179
224	allocation	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 10:59:28.628028	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "7a73f786-accc-4a9b-a2f6-7f68e47de16f", "sales_order_id": "5a6f336d-d29c-45b1-955e-e19af7441324"}	2026-07-30 10:59:28.628028
225	sale	sales_order	\N	34	\N	1	\N	\N	\N	3.000	11	\N	\N	2026-07-30 10:59:32.988679	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "b979d04d-d17a-41e2-b79c-0f52ac94db89", "sales_order_id": "5a6f336d-d29c-45b1-955e-e19af7441324", "sales_order_number": "ORD-20260730105928"}	2026-07-30 10:59:32.988679
226	sale	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 10:59:32.988679	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "7a73f786-accc-4a9b-a2f6-7f68e47de16f", "sales_order_id": "5a6f336d-d29c-45b1-955e-e19af7441324", "sales_order_number": "ORD-20260730105928"}	2026-07-30 10:59:32.988679
227	allocation	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 11:04:16.897011	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "2a27259e-9098-4d66-b3c7-1e8f868c5dfc", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26"}	2026-07-30 11:04:16.897011
228	allocation	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-30 11:04:16.906626	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "0d85453a-0c66-467d-af2a-352aae3bc1fd", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26"}	2026-07-30 11:04:16.906626
229	allocation	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 11:04:16.909565	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "d99730f8-3cfd-4c79-a4c6-c86885799d02", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26"}	2026-07-30 11:04:16.909565
230	allocation	sales_order	\N	26	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-30 11:04:16.912537	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "9cffff6b-01d3-4517-a59d-19ac4afc19f3", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26"}	2026-07-30 11:04:16.912537
231	allocation	sales_order	\N	6	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 11:04:16.915719	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "7e818c00-2787-40e9-892a-1860ee3970c7", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26"}	2026-07-30 11:04:16.915719
232	allocation	sales_order	\N	7	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 11:04:16.919066	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "ad8678ad-1ff1-4448-a833-5cd42f59fb17", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26"}	2026-07-30 11:04:16.919066
233	allocation	sales_order	\N	23	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:04:16.922166	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "089331bc-05b8-4ce9-8b9b-86ebb75bf549", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26"}	2026-07-30 11:04:16.922166
234	allocation	sales_order	\N	22	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:04:16.924423	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "13b5466e-81a0-42ff-8266-644b46c7855f", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26"}	2026-07-30 11:04:16.924423
235	allocation	sales_order	\N	40	\N	1	\N	\N	\N	1.000	4	\N	\N	2026-07-30 11:04:16.926767	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "fdf1fd54-1951-427e-8402-934ecdfe0d05", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26"}	2026-07-30 11:04:16.926767
236	allocation	sales_order	\N	41	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-30 11:04:16.930573	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "33099cdb-8963-4848-b17c-3cdad8c8f017", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26"}	2026-07-30 11:04:16.930573
237	allocation	sales_order	\N	8	\N	1	\N	\N	\N	1.000	1	\N	\N	2026-07-30 11:04:16.933065	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "70438b60-01af-4dd4-8c0d-4ebe267aeb78", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26"}	2026-07-30 11:04:16.933065
238	allocation	sales_order	\N	2	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:04:16.93515	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "b4f519de-c778-408d-84a6-69d0adf43469", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26"}	2026-07-30 11:04:16.93515
239	allocation	sales_order	\N	1	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:04:16.937371	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "d900d213-c385-4f38-bba4-60b0413f9e73", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26"}	2026-07-30 11:04:16.937371
240	allocation	sales_order	\N	32	\N	1	\N	\N	\N	1.000	2	\N	\N	2026-07-30 11:04:16.939624	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "491c0128-28ad-4172-af3e-a35814e5dfa0", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26"}	2026-07-30 11:04:16.939624
241	allocation	sales_order	\N	3	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:04:16.941683	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "35271ce0-b5c7-4209-8e75-b615cbf94a4f", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26"}	2026-07-30 11:04:16.941683
242	allocation	sales_order	\N	30	\N	1	\N	\N	\N	1.000	2	\N	\N	2026-07-30 11:04:16.943773	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "f604919b-99c7-4695-b5c7-6ff75f4a3ec5", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26"}	2026-07-30 11:04:16.943773
243	allocation	sales_order	\N	31	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 11:04:16.946198	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "a884175a-c38f-44c0-9e6e-b583a23d0c40", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26"}	2026-07-30 11:04:16.946198
244	allocation	sales_order	\N	14	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:04:16.948328	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "284f17a3-b852-4d75-9678-4d323cdbff53", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26"}	2026-07-30 11:04:16.948328
245	allocation	sales_order	\N	15	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:04:16.950519	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "dc305e0f-ee56-4977-9f01-d7301ef22f86", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26"}	2026-07-30 11:04:16.950519
246	allocation	sales_order	\N	38	\N	1	\N	\N	\N	1.000	2	\N	\N	2026-07-30 11:04:16.952811	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "880a2014-f30d-462e-82b7-ae483a25c835", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26"}	2026-07-30 11:04:16.952811
247	allocation	sales_order	\N	39	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:04:16.954966	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "eb61297c-ddcb-457c-882b-6aebbcbae1f8", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26"}	2026-07-30 11:04:16.954966
248	allocation	sales_order	\N	36	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-30 11:04:16.957231	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "de6f6b7b-f7ab-4651-8f1a-abbf2301bcba", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26"}	2026-07-30 11:04:16.957231
249	allocation	sales_order	\N	37	\N	1	\N	\N	\N	1.000	2	\N	\N	2026-07-30 11:04:16.959497	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "9d39afcf-7de8-4fd1-9dfa-382a26ef4f6c", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26"}	2026-07-30 11:04:16.959497
250	allocation	sales_order	\N	25	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 11:04:16.961691	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "db67e370-d33b-4d2f-953e-a1e600a95c3d", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26"}	2026-07-30 11:04:16.961691
251	allocation	sales_order	\N	20	\N	1	\N	\N	\N	1.000	2	\N	\N	2026-07-30 11:04:16.96385	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "29af91ac-b256-4cc0-b80f-ebea44c0fe52", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26"}	2026-07-30 11:04:16.96385
252	allocation	sales_order	\N	21	\N	1	\N	\N	\N	1.000	2	\N	\N	2026-07-30 11:04:16.966227	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "c48fe4d0-f363-483d-8291-5f964dcbb995", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26"}	2026-07-30 11:04:16.966227
253	allocation	sales_order	\N	11	\N	1	\N	\N	\N	1.000	7	\N	\N	2026-07-30 11:04:16.968346	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "d1f4dee4-218f-46ce-8d54-037a2e3b3bf9", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26"}	2026-07-30 11:04:16.968346
254	allocation	sales_order	\N	9	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-30 11:04:16.970614	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "3ec3c2fe-66cd-4dfa-a793-6d948e1e09d8", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26"}	2026-07-30 11:04:16.970614
255	sale	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 11:04:24.984511	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "2a27259e-9098-4d66-b3c7-1e8f868c5dfc", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26", "sales_order_number": "ORD-20260730110416"}	2026-07-30 11:04:24.984511
256	sale	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-30 11:04:24.984511	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "0d85453a-0c66-467d-af2a-352aae3bc1fd", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26", "sales_order_number": "ORD-20260730110416"}	2026-07-30 11:04:24.984511
257	sale	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 11:04:24.984511	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "d99730f8-3cfd-4c79-a4c6-c86885799d02", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26", "sales_order_number": "ORD-20260730110416"}	2026-07-30 11:04:24.984511
258	sale	sales_order	\N	26	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-30 11:04:24.984511	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "9cffff6b-01d3-4517-a59d-19ac4afc19f3", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26", "sales_order_number": "ORD-20260730110416"}	2026-07-30 11:04:24.984511
456	transfer_in	transfer_request	14	9	\N	4	1	9	2	2.000	5	\N	\N	2026-08-11 07:33:05.439702	2	completed	\N	\N	{"transfer_number": "TRF-COKE-00454"}	2026-08-11 07:33:05.439702
259	sale	sales_order	\N	6	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 11:04:24.984511	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "7e818c00-2787-40e9-892a-1860ee3970c7", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26", "sales_order_number": "ORD-20260730110416"}	2026-07-30 11:04:24.984511
260	sale	sales_order	\N	7	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 11:04:24.984511	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "ad8678ad-1ff1-4448-a833-5cd42f59fb17", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26", "sales_order_number": "ORD-20260730110416"}	2026-07-30 11:04:24.984511
261	sale	sales_order	\N	23	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:04:24.984511	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "089331bc-05b8-4ce9-8b9b-86ebb75bf549", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26", "sales_order_number": "ORD-20260730110416"}	2026-07-30 11:04:24.984511
262	sale	sales_order	\N	22	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:04:24.984511	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "13b5466e-81a0-42ff-8266-644b46c7855f", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26", "sales_order_number": "ORD-20260730110416"}	2026-07-30 11:04:24.984511
263	sale	sales_order	\N	40	\N	1	\N	\N	\N	1.000	4	\N	\N	2026-07-30 11:04:24.984511	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "fdf1fd54-1951-427e-8402-934ecdfe0d05", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26", "sales_order_number": "ORD-20260730110416"}	2026-07-30 11:04:24.984511
264	sale	sales_order	\N	41	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-30 11:04:24.984511	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "33099cdb-8963-4848-b17c-3cdad8c8f017", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26", "sales_order_number": "ORD-20260730110416"}	2026-07-30 11:04:24.984511
265	sale	sales_order	\N	8	\N	1	\N	\N	\N	1.000	1	\N	\N	2026-07-30 11:04:24.984511	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "70438b60-01af-4dd4-8c0d-4ebe267aeb78", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26", "sales_order_number": "ORD-20260730110416"}	2026-07-30 11:04:24.984511
266	sale	sales_order	\N	2	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:04:24.984511	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "b4f519de-c778-408d-84a6-69d0adf43469", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26", "sales_order_number": "ORD-20260730110416"}	2026-07-30 11:04:24.984511
267	sale	sales_order	\N	1	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:04:24.984511	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "d900d213-c385-4f38-bba4-60b0413f9e73", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26", "sales_order_number": "ORD-20260730110416"}	2026-07-30 11:04:24.984511
268	sale	sales_order	\N	32	\N	1	\N	\N	\N	1.000	2	\N	\N	2026-07-30 11:04:24.984511	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "491c0128-28ad-4172-af3e-a35814e5dfa0", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26", "sales_order_number": "ORD-20260730110416"}	2026-07-30 11:04:24.984511
269	sale	sales_order	\N	3	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:04:24.984511	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "35271ce0-b5c7-4209-8e75-b615cbf94a4f", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26", "sales_order_number": "ORD-20260730110416"}	2026-07-30 11:04:24.984511
270	sale	sales_order	\N	30	\N	1	\N	\N	\N	1.000	2	\N	\N	2026-07-30 11:04:24.984511	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "f604919b-99c7-4695-b5c7-6ff75f4a3ec5", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26", "sales_order_number": "ORD-20260730110416"}	2026-07-30 11:04:24.984511
271	sale	sales_order	\N	31	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 11:04:24.984511	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "a884175a-c38f-44c0-9e6e-b583a23d0c40", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26", "sales_order_number": "ORD-20260730110416"}	2026-07-30 11:04:24.984511
272	sale	sales_order	\N	14	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:04:24.984511	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "284f17a3-b852-4d75-9678-4d323cdbff53", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26", "sales_order_number": "ORD-20260730110416"}	2026-07-30 11:04:24.984511
273	sale	sales_order	\N	15	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:04:24.984511	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "dc305e0f-ee56-4977-9f01-d7301ef22f86", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26", "sales_order_number": "ORD-20260730110416"}	2026-07-30 11:04:24.984511
274	sale	sales_order	\N	38	\N	1	\N	\N	\N	1.000	2	\N	\N	2026-07-30 11:04:24.984511	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "880a2014-f30d-462e-82b7-ae483a25c835", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26", "sales_order_number": "ORD-20260730110416"}	2026-07-30 11:04:24.984511
275	sale	sales_order	\N	39	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:04:24.984511	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "eb61297c-ddcb-457c-882b-6aebbcbae1f8", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26", "sales_order_number": "ORD-20260730110416"}	2026-07-30 11:04:24.984511
276	sale	sales_order	\N	36	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-30 11:04:24.984511	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "de6f6b7b-f7ab-4651-8f1a-abbf2301bcba", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26", "sales_order_number": "ORD-20260730110416"}	2026-07-30 11:04:24.984511
277	sale	sales_order	\N	37	\N	1	\N	\N	\N	1.000	2	\N	\N	2026-07-30 11:04:24.984511	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "9d39afcf-7de8-4fd1-9dfa-382a26ef4f6c", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26", "sales_order_number": "ORD-20260730110416"}	2026-07-30 11:04:24.984511
278	sale	sales_order	\N	25	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 11:04:24.984511	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "db67e370-d33b-4d2f-953e-a1e600a95c3d", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26", "sales_order_number": "ORD-20260730110416"}	2026-07-30 11:04:24.984511
279	sale	sales_order	\N	20	\N	1	\N	\N	\N	1.000	2	\N	\N	2026-07-30 11:04:24.984511	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "29af91ac-b256-4cc0-b80f-ebea44c0fe52", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26", "sales_order_number": "ORD-20260730110416"}	2026-07-30 11:04:24.984511
280	sale	sales_order	\N	21	\N	1	\N	\N	\N	1.000	2	\N	\N	2026-07-30 11:04:24.984511	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "c48fe4d0-f363-483d-8291-5f964dcbb995", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26", "sales_order_number": "ORD-20260730110416"}	2026-07-30 11:04:24.984511
281	sale	sales_order	\N	11	\N	1	\N	\N	\N	1.000	7	\N	\N	2026-07-30 11:04:24.984511	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "d1f4dee4-218f-46ce-8d54-037a2e3b3bf9", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26", "sales_order_number": "ORD-20260730110416"}	2026-07-30 11:04:24.984511
282	sale	sales_order	\N	9	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-30 11:04:24.984511	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "3ec3c2fe-66cd-4dfa-a793-6d948e1e09d8", "sales_order_id": "a4e64219-9b7f-45c4-9a35-106788960f26", "sales_order_number": "ORD-20260730110416"}	2026-07-30 11:04:24.984511
283	allocation	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 11:08:54.541481	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "384cb23e-ce0e-4472-82c0-7e5331f3c42f", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.541481
284	allocation	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-30 11:08:54.550108	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "d0f5f849-5266-47b3-b76e-52ae5c20d5f4", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.550108
457	transfer_out	transfer_request	15	9	\N	4	1	9	2	2.000	5	\N	\N	2026-08-11 07:33:59.170837	1	shipped	\N	\N	{"transfer_number": "TRF-COKE-00457"}	2026-08-11 07:33:59.170837
285	allocation	sales_order	\N	26	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-30 11:08:54.552957	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "8917fd7f-717e-49c7-9f19-f3aa2e4b732a", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.552957
286	allocation	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 11:08:54.555775	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "21069664-16d4-456f-ab14-5acffee00fe3", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.555775
287	allocation	sales_order	\N	27	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-30 11:08:54.558792	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "049e278a-7f8d-47f2-bcb0-01e872b02e86", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.558792
288	allocation	sales_order	\N	6	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 11:08:54.561753	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "737dde08-8d0c-4262-a5d8-ee6da5249a99", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.561753
289	allocation	sales_order	\N	23	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:08:54.564857	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "5e4e3063-b537-4f73-86cd-ffdf566726ce", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.564857
290	allocation	sales_order	\N	7	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 11:08:54.567133	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "5f58e161-360b-4640-bb8b-061648468986", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.567133
291	allocation	sales_order	\N	22	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:08:54.569248	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "bb4de653-6d5f-4cf3-be52-70862402cf11", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.569248
292	allocation	sales_order	\N	40	\N	1	\N	\N	\N	1.000	4	\N	\N	2026-07-30 11:08:54.571371	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "eb3ea546-9a71-4274-9abd-c98ba7b0a7c1", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.571371
293	allocation	sales_order	\N	8	\N	1	\N	\N	\N	1.000	1	\N	\N	2026-07-30 11:08:54.573522	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "009d75c0-7257-4851-a5c7-e488a5dade74", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.573522
294	allocation	sales_order	\N	41	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-30 11:08:54.575606	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "c51ce895-4af8-42f9-88d8-7057f4ff5c8a", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.575606
295	allocation	sales_order	\N	1	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:08:54.577699	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "a79fcb7d-1b83-434a-b0aa-21796c072e7c", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.577699
296	allocation	sales_order	\N	2	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:08:54.579897	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "26426806-04df-467f-a382-ee3e16df8e4c", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.579897
297	allocation	sales_order	\N	32	\N	1	\N	\N	\N	1.000	2	\N	\N	2026-07-30 11:08:54.582048	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "928e9ea3-ddbc-40d6-b046-9c8e4929774f", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.582048
298	allocation	sales_order	\N	3	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:08:54.584129	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "19b554c8-d8ca-46ce-bdf9-2d22ff40b0e3", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.584129
299	allocation	sales_order	\N	30	\N	1	\N	\N	\N	1.000	2	\N	\N	2026-07-30 11:08:54.586501	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "65a4a445-8b20-4d2b-ad3e-24c11097fcdc", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.586501
300	allocation	sales_order	\N	31	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 11:08:54.588633	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "8cba9e06-2d83-4c6b-a1db-44db06cde508", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.588633
301	allocation	sales_order	\N	14	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:08:54.590771	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "2a495556-3ab7-4985-a173-24b96187a6d9", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.590771
302	allocation	sales_order	\N	15	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:08:54.593087	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "b8687304-98c5-4355-8a4c-b621bb7212fe", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.593087
303	allocation	sales_order	\N	39	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:08:54.595234	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "e2344d7d-d9e6-424a-aa79-9b5d712deed2", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.595234
304	allocation	sales_order	\N	38	\N	1	\N	\N	\N	1.000	2	\N	\N	2026-07-30 11:08:54.597544	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "c2e7db52-8f44-4d98-a0f9-729d04cffe95", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.597544
305	allocation	sales_order	\N	36	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-30 11:08:54.599737	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "7665d1b8-5343-466e-88c4-c60a622b9c00", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.599737
306	allocation	sales_order	\N	25	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 11:08:54.601872	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "daa2475d-bd25-47e8-9391-ae186ba3df51", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.601872
307	allocation	sales_order	\N	24	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 11:08:54.604008	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "98b48bcf-d29e-4c73-ab0a-7b051c32affa", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.604008
308	allocation	sales_order	\N	20	\N	1	\N	\N	\N	1.000	2	\N	\N	2026-07-30 11:08:54.606333	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "c755ef48-a864-421c-bbde-508777c469aa", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.606333
309	allocation	sales_order	\N	11	\N	1	\N	\N	\N	1.000	7	\N	\N	2026-07-30 11:08:54.608483	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "119b450c-f641-419b-b23d-e87993a96c18", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.608483
310	allocation	sales_order	\N	29	\N	1	\N	\N	\N	1.000	2	\N	\N	2026-07-30 11:08:54.610685	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "c71e7b1f-7791-4840-bcd1-09e69eb6910b", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.610685
311	allocation	sales_order	\N	10	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-30 11:08:54.612919	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "9201e228-d832-4966-9f3b-883ea7e51fc4", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.612919
312	allocation	sales_order	\N	17	\N	1	\N	\N	\N	1.000	4	\N	\N	2026-07-30 11:08:54.615004	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "c8375161-8569-4977-a53d-60ee355e3f6e", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.615004
313	allocation	sales_order	\N	19	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 11:08:54.617093	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "2f879176-9d33-41ca-b783-9ff6bd6e92b5", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.617093
314	allocation	sales_order	\N	16	\N	1	\N	\N	\N	1.000	4	\N	\N	2026-07-30 11:08:54.619169	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "96829edd-27e6-4858-894a-a55e9998ebd4", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.619169
315	allocation	sales_order	\N	13	\N	1	\N	\N	\N	1.000	7	\N	\N	2026-07-30 11:08:54.621258	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "e21c434f-e190-4627-8c4d-3cc8c6d203c0", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.621258
316	allocation	sales_order	\N	5	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 11:08:54.623362	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "e707c558-499a-4d08-ab54-06068f2f7f59", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8"}	2026-07-30 11:08:54.623362
317	sale	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "384cb23e-ce0e-4472-82c0-7e5331f3c42f", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
318	sale	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "d0f5f849-5266-47b3-b76e-52ae5c20d5f4", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
319	sale	sales_order	\N	26	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "8917fd7f-717e-49c7-9f19-f3aa2e4b732a", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
320	sale	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "21069664-16d4-456f-ab14-5acffee00fe3", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
321	sale	sales_order	\N	27	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "049e278a-7f8d-47f2-bcb0-01e872b02e86", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
322	sale	sales_order	\N	6	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "737dde08-8d0c-4262-a5d8-ee6da5249a99", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
323	sale	sales_order	\N	23	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "5e4e3063-b537-4f73-86cd-ffdf566726ce", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
324	sale	sales_order	\N	7	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "5f58e161-360b-4640-bb8b-061648468986", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
325	sale	sales_order	\N	22	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "bb4de653-6d5f-4cf3-be52-70862402cf11", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
326	sale	sales_order	\N	40	\N	1	\N	\N	\N	1.000	4	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "eb3ea546-9a71-4274-9abd-c98ba7b0a7c1", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
327	sale	sales_order	\N	8	\N	1	\N	\N	\N	1.000	1	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "009d75c0-7257-4851-a5c7-e488a5dade74", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
328	sale	sales_order	\N	41	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "c51ce895-4af8-42f9-88d8-7057f4ff5c8a", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
329	sale	sales_order	\N	1	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "a79fcb7d-1b83-434a-b0aa-21796c072e7c", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
330	sale	sales_order	\N	2	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "26426806-04df-467f-a382-ee3e16df8e4c", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
331	sale	sales_order	\N	32	\N	1	\N	\N	\N	1.000	2	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "928e9ea3-ddbc-40d6-b046-9c8e4929774f", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
332	sale	sales_order	\N	3	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "19b554c8-d8ca-46ce-bdf9-2d22ff40b0e3", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
333	sale	sales_order	\N	30	\N	1	\N	\N	\N	1.000	2	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "65a4a445-8b20-4d2b-ad3e-24c11097fcdc", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
334	sale	sales_order	\N	31	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "8cba9e06-2d83-4c6b-a1db-44db06cde508", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
335	sale	sales_order	\N	14	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "2a495556-3ab7-4985-a173-24b96187a6d9", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
336	sale	sales_order	\N	15	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "b8687304-98c5-4355-8a4c-b621bb7212fe", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
337	sale	sales_order	\N	39	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "e2344d7d-d9e6-424a-aa79-9b5d712deed2", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
338	sale	sales_order	\N	38	\N	1	\N	\N	\N	1.000	2	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "c2e7db52-8f44-4d98-a0f9-729d04cffe95", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
339	sale	sales_order	\N	36	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "7665d1b8-5343-466e-88c4-c60a622b9c00", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
340	sale	sales_order	\N	25	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "daa2475d-bd25-47e8-9391-ae186ba3df51", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
458	transfer_in	transfer_request	15	9	\N	4	1	9	2	2.000	5	\N	\N	2026-08-11 07:34:08.549366	2	completed	\N	\N	{"transfer_number": "TRF-COKE-00457"}	2026-08-11 07:34:08.549366
341	sale	sales_order	\N	24	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "98b48bcf-d29e-4c73-ab0a-7b051c32affa", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
342	sale	sales_order	\N	20	\N	1	\N	\N	\N	1.000	2	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "c755ef48-a864-421c-bbde-508777c469aa", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
343	sale	sales_order	\N	11	\N	1	\N	\N	\N	1.000	7	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "119b450c-f641-419b-b23d-e87993a96c18", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
344	sale	sales_order	\N	29	\N	1	\N	\N	\N	1.000	2	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "c71e7b1f-7791-4840-bcd1-09e69eb6910b", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
345	sale	sales_order	\N	10	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "9201e228-d832-4966-9f3b-883ea7e51fc4", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
346	sale	sales_order	\N	17	\N	1	\N	\N	\N	1.000	4	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "c8375161-8569-4977-a53d-60ee355e3f6e", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
347	sale	sales_order	\N	19	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "2f879176-9d33-41ca-b783-9ff6bd6e92b5", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
348	sale	sales_order	\N	16	\N	1	\N	\N	\N	1.000	4	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "96829edd-27e6-4858-894a-a55e9998ebd4", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
349	sale	sales_order	\N	13	\N	1	\N	\N	\N	1.000	7	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "e21c434f-e190-4627-8c4d-3cc8c6d203c0", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
350	sale	sales_order	\N	5	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-30 11:08:58.292688	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "e707c558-499a-4d08-ab54-06068f2f7f59", "sales_order_id": "54bbfac0-7b51-41f2-a2a5-bfcd7199c5b8", "sales_order_number": "ORD-20260730110854"}	2026-07-30 11:08:58.292688
351	transfer_out	transfer_request	9	34	\N	1	17	7	12	50.000	11	\N	\N	2026-07-31 06:17:59.038853	1	shipped	\N	\N	{"transfer_number": "TRF-58E2-4DC3"}	2026-07-31 06:17:59.038853
352	transfer_in	transfer_request	9	34	\N	1	17	7	12	50.000	11	\N	\N	2026-07-31 06:18:20.713763	1	completed	\N	\N	{"transfer_number": "TRF-58E2-4DC3"}	2026-07-31 06:18:20.713763
353	allocation	sales_order	\N	5	\N	8	\N	\N	\N	1.000	10	\N	\N	2026-07-31 09:02:44.042278	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "e886b0ef-2f47-4c2a-92a5-6d0ee4fbf368", "sales_order_id": "d7b077c6-d905-4a48-b75d-96acf66acca6"}	2026-07-31 09:02:44.042278
354	allocation	sales_order	\N	5	\N	8	\N	\N	\N	1.000	10	\N	\N	2026-07-31 09:02:44.052183	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "06dffbce-3c70-42d6-ab1b-612438e17d9a", "sales_order_id": "d7b077c6-d905-4a48-b75d-96acf66acca6"}	2026-07-31 09:02:44.052183
355	allocation	sales_order	\N	8	\N	8	\N	\N	\N	1.000	1	\N	\N	2026-07-31 09:02:44.055243	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "8f85143e-b0be-454f-b9eb-3a3f9e405548", "sales_order_id": "d7b077c6-d905-4a48-b75d-96acf66acca6"}	2026-07-31 09:02:44.055243
356	allocation	sales_order	\N	5	\N	8	\N	\N	\N	1.000	10	\N	\N	2026-07-31 09:02:44.058468	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "cfa84208-696e-43d6-b5f7-f2c2567e7c96", "sales_order_id": "d7b077c6-d905-4a48-b75d-96acf66acca6"}	2026-07-31 09:02:44.058468
357	allocation	sales_order	\N	7	\N	8	\N	\N	\N	1.000	10	\N	\N	2026-07-31 09:02:44.061595	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "8855d957-9652-4729-abd3-d60f307eee99", "sales_order_id": "d7b077c6-d905-4a48-b75d-96acf66acca6"}	2026-07-31 09:02:44.061595
358	allocation	sales_order	\N	5	\N	8	\N	\N	\N	1.000	10	\N	\N	2026-07-31 09:02:44.06466	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "d3356493-0292-4b3a-bf7f-590a6573b7df", "sales_order_id": "d7b077c6-d905-4a48-b75d-96acf66acca6"}	2026-07-31 09:02:44.06466
359	allocation	sales_order	\N	5	\N	8	\N	\N	\N	1.000	10	\N	\N	2026-07-31 09:02:44.067823	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "8be50779-529a-465f-9c12-9d7a73a4086c", "sales_order_id": "d7b077c6-d905-4a48-b75d-96acf66acca6"}	2026-07-31 09:02:44.067823
360	allocation	sales_order	\N	5	\N	8	\N	\N	\N	1.000	10	\N	\N	2026-07-31 09:02:44.070053	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "ed7d5145-52d4-4b55-9c85-de34b061a4f2", "sales_order_id": "d7b077c6-d905-4a48-b75d-96acf66acca6"}	2026-07-31 09:02:44.070053
361	allocation	sales_order	\N	5	\N	8	\N	\N	\N	1.000	10	\N	\N	2026-07-31 09:02:44.072392	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "af52c279-a3ee-44ac-b31d-142d57ceffe7", "sales_order_id": "d7b077c6-d905-4a48-b75d-96acf66acca6"}	2026-07-31 09:02:44.072392
362	sale	sales_order	\N	5	\N	8	\N	\N	\N	1.000	10	\N	\N	2026-07-31 09:02:52.507259	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "e886b0ef-2f47-4c2a-92a5-6d0ee4fbf368", "sales_order_id": "d7b077c6-d905-4a48-b75d-96acf66acca6", "sales_order_number": "ORD-20260731090244"}	2026-07-31 09:02:52.507259
363	sale	sales_order	\N	5	\N	8	\N	\N	\N	1.000	10	\N	\N	2026-07-31 09:02:52.507259	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "06dffbce-3c70-42d6-ab1b-612438e17d9a", "sales_order_id": "d7b077c6-d905-4a48-b75d-96acf66acca6", "sales_order_number": "ORD-20260731090244"}	2026-07-31 09:02:52.507259
364	sale	sales_order	\N	8	\N	8	\N	\N	\N	1.000	1	\N	\N	2026-07-31 09:02:52.507259	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "8f85143e-b0be-454f-b9eb-3a3f9e405548", "sales_order_id": "d7b077c6-d905-4a48-b75d-96acf66acca6", "sales_order_number": "ORD-20260731090244"}	2026-07-31 09:02:52.507259
365	sale	sales_order	\N	5	\N	8	\N	\N	\N	1.000	10	\N	\N	2026-07-31 09:02:52.507259	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "cfa84208-696e-43d6-b5f7-f2c2567e7c96", "sales_order_id": "d7b077c6-d905-4a48-b75d-96acf66acca6", "sales_order_number": "ORD-20260731090244"}	2026-07-31 09:02:52.507259
366	sale	sales_order	\N	7	\N	8	\N	\N	\N	1.000	10	\N	\N	2026-07-31 09:02:52.507259	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "8855d957-9652-4729-abd3-d60f307eee99", "sales_order_id": "d7b077c6-d905-4a48-b75d-96acf66acca6", "sales_order_number": "ORD-20260731090244"}	2026-07-31 09:02:52.507259
367	sale	sales_order	\N	5	\N	8	\N	\N	\N	1.000	10	\N	\N	2026-07-31 09:02:52.507259	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "d3356493-0292-4b3a-bf7f-590a6573b7df", "sales_order_id": "d7b077c6-d905-4a48-b75d-96acf66acca6", "sales_order_number": "ORD-20260731090244"}	2026-07-31 09:02:52.507259
368	sale	sales_order	\N	5	\N	8	\N	\N	\N	1.000	10	\N	\N	2026-07-31 09:02:52.507259	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "8be50779-529a-465f-9c12-9d7a73a4086c", "sales_order_id": "d7b077c6-d905-4a48-b75d-96acf66acca6", "sales_order_number": "ORD-20260731090244"}	2026-07-31 09:02:52.507259
369	sale	sales_order	\N	5	\N	8	\N	\N	\N	1.000	10	\N	\N	2026-07-31 09:02:52.507259	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "ed7d5145-52d4-4b55-9c85-de34b061a4f2", "sales_order_id": "d7b077c6-d905-4a48-b75d-96acf66acca6", "sales_order_number": "ORD-20260731090244"}	2026-07-31 09:02:52.507259
370	sale	sales_order	\N	5	\N	8	\N	\N	\N	1.000	10	\N	\N	2026-07-31 09:02:52.507259	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "af52c279-a3ee-44ac-b31d-142d57ceffe7", "sales_order_id": "d7b077c6-d905-4a48-b75d-96acf66acca6", "sales_order_number": "ORD-20260731090244"}	2026-07-31 09:02:52.507259
371	allocation	sales_order	\N	5	\N	8	\N	\N	\N	1.000	10	\N	\N	2026-07-31 09:05:05.116104	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "c10fbda6-83e7-4806-8743-daa1ced8c071", "sales_order_id": "bcf0f68b-1976-4186-9409-13d858a952f6"}	2026-07-31 09:05:05.116104
372	sale	sales_order	\N	5	\N	8	\N	\N	\N	1.000	10	\N	\N	2026-07-31 09:05:23.634657	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "c10fbda6-83e7-4806-8743-daa1ced8c071", "sales_order_id": "bcf0f68b-1976-4186-9409-13d858a952f6", "sales_order_number": "ORD-20260731090505"}	2026-07-31 09:05:23.634657
373	allocation	sales_order	\N	33	\N	1	\N	\N	\N	3.000	10	\N	\N	2026-07-31 09:11:02.471786	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "239d2405-bed6-4ed1-b6d3-abbd12092f67", "sales_order_id": "e6c4a620-4b91-4708-bec4-40e04779346b"}	2026-07-31 09:11:02.471786
374	allocation	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-31 09:11:02.47625	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "cc919936-5156-4352-95a1-11e25eae1308", "sales_order_id": "e6c4a620-4b91-4708-bec4-40e04779346b"}	2026-07-31 09:11:02.47625
375	allocation	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-31 09:11:02.478966	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "02237dbc-8078-47e5-8ce6-7ad11759a4ba", "sales_order_id": "e6c4a620-4b91-4708-bec4-40e04779346b"}	2026-07-31 09:11:02.478966
376	allocation	sales_order	\N	26	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-31 09:11:02.481656	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "81482efb-f31e-4ea1-827c-2ab8c09ef8e0", "sales_order_id": "e6c4a620-4b91-4708-bec4-40e04779346b"}	2026-07-31 09:11:02.481656
377	allocation	sales_order	\N	27	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-31 09:11:02.484246	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "5d63cbc7-1091-43ea-a9e5-31aa0a6a8f56", "sales_order_id": "e6c4a620-4b91-4708-bec4-40e04779346b"}	2026-07-31 09:11:02.484246
378	allocation	sales_order	\N	2	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-31 09:11:02.487002	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "c552ab75-409b-4d48-98b5-aee2a34720e9", "sales_order_id": "e6c4a620-4b91-4708-bec4-40e04779346b"}	2026-07-31 09:11:02.487002
379	sale	sales_order	\N	33	\N	1	\N	\N	\N	3.000	10	\N	\N	2026-07-31 09:11:07.943889	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "239d2405-bed6-4ed1-b6d3-abbd12092f67", "sales_order_id": "e6c4a620-4b91-4708-bec4-40e04779346b", "sales_order_number": "ORD-20260731091102"}	2026-07-31 09:11:07.943889
380	sale	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-31 09:11:07.943889	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "cc919936-5156-4352-95a1-11e25eae1308", "sales_order_id": "e6c4a620-4b91-4708-bec4-40e04779346b", "sales_order_number": "ORD-20260731091102"}	2026-07-31 09:11:07.943889
381	sale	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-31 09:11:07.943889	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "02237dbc-8078-47e5-8ce6-7ad11759a4ba", "sales_order_id": "e6c4a620-4b91-4708-bec4-40e04779346b", "sales_order_number": "ORD-20260731091102"}	2026-07-31 09:11:07.943889
382	sale	sales_order	\N	26	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-31 09:11:07.943889	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "81482efb-f31e-4ea1-827c-2ab8c09ef8e0", "sales_order_id": "e6c4a620-4b91-4708-bec4-40e04779346b", "sales_order_number": "ORD-20260731091102"}	2026-07-31 09:11:07.943889
383	sale	sales_order	\N	27	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-07-31 09:11:07.943889	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "5d63cbc7-1091-43ea-a9e5-31aa0a6a8f56", "sales_order_id": "e6c4a620-4b91-4708-bec4-40e04779346b", "sales_order_number": "ORD-20260731091102"}	2026-07-31 09:11:07.943889
384	sale	sales_order	\N	2	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-31 09:11:07.943889	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "c552ab75-409b-4d48-98b5-aee2a34720e9", "sales_order_id": "e6c4a620-4b91-4708-bec4-40e04779346b", "sales_order_number": "ORD-20260731091102"}	2026-07-31 09:11:07.943889
385	allocation	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-31 09:13:36.24426	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "146a4b84-a25c-4fd8-8538-b789be1d8b8b", "sales_order_id": "0fcd3efb-e505-45fe-b224-b8010407fd0e"}	2026-07-31 09:13:36.24426
386	allocation	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-31 09:13:36.247388	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "a9abf26c-ea84-4588-a9a4-ed220794e9a4", "sales_order_id": "0fcd3efb-e505-45fe-b224-b8010407fd0e"}	2026-07-31 09:13:36.247388
387	sale	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-31 09:13:42.24131	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "146a4b84-a25c-4fd8-8538-b789be1d8b8b", "sales_order_id": "0fcd3efb-e505-45fe-b224-b8010407fd0e", "sales_order_number": "ORD-20260731091336"}	2026-07-31 09:13:42.24131
388	sale	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-31 09:13:42.24131	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "a9abf26c-ea84-4588-a9a4-ed220794e9a4", "sales_order_id": "0fcd3efb-e505-45fe-b224-b8010407fd0e", "sales_order_number": "ORD-20260731091336"}	2026-07-31 09:13:42.24131
389	allocation	sales_order	\N	33	\N	1	\N	\N	\N	2.000	10	\N	\N	2026-07-31 10:19:10.727492	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "5c369556-2879-4306-8818-99a8ea0d8574", "sales_order_id": "8076b3b4-166f-457d-8358-ba2fd13a3832"}	2026-07-31 10:19:10.727492
390	allocation	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-31 10:19:10.735929	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "589dde9b-04d1-471e-bcf8-cb4db7cd0a56", "sales_order_id": "8076b3b4-166f-457d-8358-ba2fd13a3832"}	2026-07-31 10:19:10.735929
391	allocation	sales_order	\N	26	\N	1	\N	\N	\N	2.000	8	\N	\N	2026-07-31 10:19:10.738843	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "6762a9ee-0690-42e4-870c-7d421e98ce67", "sales_order_id": "8076b3b4-166f-457d-8358-ba2fd13a3832"}	2026-07-31 10:19:10.738843
392	allocation	sales_order	\N	6	\N	1	\N	\N	\N	2.000	10	\N	\N	2026-07-31 10:19:10.741953	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "aa2d1aad-1bd8-479f-a3fa-ed65e40d8ba5", "sales_order_id": "8076b3b4-166f-457d-8358-ba2fd13a3832"}	2026-07-31 10:19:10.741953
393	allocation	sales_order	\N	3	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-31 10:19:10.745015	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "648cca87-c158-41b4-9c14-542a57e1645d", "sales_order_id": "8076b3b4-166f-457d-8358-ba2fd13a3832"}	2026-07-31 10:19:10.745015
394	allocation	sales_order	\N	31	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-31 10:19:10.747953	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "51e64007-67ef-468e-8329-5ae036742ffb", "sales_order_id": "8076b3b4-166f-457d-8358-ba2fd13a3832"}	2026-07-31 10:19:10.747953
395	allocation	sales_order	\N	30	\N	1	\N	\N	\N	2.000	2	\N	\N	2026-07-31 10:19:10.751102	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "d1ed4f45-feb1-4d67-acc5-1da79255a70d", "sales_order_id": "8076b3b4-166f-457d-8358-ba2fd13a3832"}	2026-07-31 10:19:10.751102
396	allocation	sales_order	\N	41	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-31 10:19:10.753317	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "1e2729e1-a7ef-46fd-b0b2-00844936e093", "sales_order_id": "8076b3b4-166f-457d-8358-ba2fd13a3832"}	2026-07-31 10:19:10.753317
459	transfer_out	transfer_request	16	9	\N	4	1	9	2	48.000	5	\N	\N	2026-08-11 07:52:25.085531	1	shipped	\N	\N	{"transfer_number": "TRF-COKE-00458"}	2026-08-11 07:52:25.085531
397	allocation	sales_order	\N	2	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-31 10:19:10.755587	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "0350c319-2184-402c-aac4-a9160e587bce", "sales_order_id": "8076b3b4-166f-457d-8358-ba2fd13a3832"}	2026-07-31 10:19:10.755587
398	allocation	sales_order	\N	37	\N	1	\N	\N	\N	2.000	2	\N	\N	2026-07-31 10:19:10.757762	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "ef314698-c881-4b6c-8398-385f942e2f37", "sales_order_id": "8076b3b4-166f-457d-8358-ba2fd13a3832"}	2026-07-31 10:19:10.757762
399	sale	sales_order	\N	33	\N	1	\N	\N	\N	2.000	10	\N	\N	2026-07-31 10:19:16.186314	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "5c369556-2879-4306-8818-99a8ea0d8574", "sales_order_id": "8076b3b4-166f-457d-8358-ba2fd13a3832", "sales_order_number": "ORD-20260731101910"}	2026-07-31 10:19:16.186314
400	sale	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-31 10:19:16.186314	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "589dde9b-04d1-471e-bcf8-cb4db7cd0a56", "sales_order_id": "8076b3b4-166f-457d-8358-ba2fd13a3832", "sales_order_number": "ORD-20260731101910"}	2026-07-31 10:19:16.186314
401	sale	sales_order	\N	26	\N	1	\N	\N	\N	2.000	8	\N	\N	2026-07-31 10:19:16.186314	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "6762a9ee-0690-42e4-870c-7d421e98ce67", "sales_order_id": "8076b3b4-166f-457d-8358-ba2fd13a3832", "sales_order_number": "ORD-20260731101910"}	2026-07-31 10:19:16.186314
402	sale	sales_order	\N	6	\N	1	\N	\N	\N	2.000	10	\N	\N	2026-07-31 10:19:16.186314	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "aa2d1aad-1bd8-479f-a3fa-ed65e40d8ba5", "sales_order_id": "8076b3b4-166f-457d-8358-ba2fd13a3832", "sales_order_number": "ORD-20260731101910"}	2026-07-31 10:19:16.186314
403	sale	sales_order	\N	3	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-31 10:19:16.186314	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "648cca87-c158-41b4-9c14-542a57e1645d", "sales_order_id": "8076b3b4-166f-457d-8358-ba2fd13a3832", "sales_order_number": "ORD-20260731101910"}	2026-07-31 10:19:16.186314
404	sale	sales_order	\N	31	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-07-31 10:19:16.186314	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "51e64007-67ef-468e-8329-5ae036742ffb", "sales_order_id": "8076b3b4-166f-457d-8358-ba2fd13a3832", "sales_order_number": "ORD-20260731101910"}	2026-07-31 10:19:16.186314
405	sale	sales_order	\N	30	\N	1	\N	\N	\N	2.000	2	\N	\N	2026-07-31 10:19:16.186314	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "d1ed4f45-feb1-4d67-acc5-1da79255a70d", "sales_order_id": "8076b3b4-166f-457d-8358-ba2fd13a3832", "sales_order_number": "ORD-20260731101910"}	2026-07-31 10:19:16.186314
406	sale	sales_order	\N	41	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-07-31 10:19:16.186314	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "1e2729e1-a7ef-46fd-b0b2-00844936e093", "sales_order_id": "8076b3b4-166f-457d-8358-ba2fd13a3832", "sales_order_number": "ORD-20260731101910"}	2026-07-31 10:19:16.186314
407	sale	sales_order	\N	2	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-07-31 10:19:16.186314	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "0350c319-2184-402c-aac4-a9160e587bce", "sales_order_id": "8076b3b4-166f-457d-8358-ba2fd13a3832", "sales_order_number": "ORD-20260731101910"}	2026-07-31 10:19:16.186314
408	sale	sales_order	\N	37	\N	1	\N	\N	\N	2.000	2	\N	\N	2026-07-31 10:19:16.186314	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "ef314698-c881-4b6c-8398-385f942e2f37", "sales_order_id": "8076b3b4-166f-457d-8358-ba2fd13a3832", "sales_order_number": "ORD-20260731101910"}	2026-07-31 10:19:16.186314
409	allocation	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-08-01 06:18:24.45555	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "0e13ac15-3459-4f35-b2e4-51d47fd7c036", "sales_order_id": "098814e2-f46e-4be6-9730-87857812e82a"}	2026-08-01 06:18:24.45555
410	allocation	sales_order	\N	22	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-08-01 06:18:24.464519	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "52151a2d-fbef-4eb6-bc80-e113d10aa481", "sales_order_id": "098814e2-f46e-4be6-9730-87857812e82a"}	2026-08-01 06:18:24.464519
411	allocation	sales_order	\N	8	\N	1	\N	\N	\N	1.000	1	\N	\N	2026-08-01 06:18:24.468094	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "83e7f800-e9c6-4962-93e4-e1acc13d0aa4", "sales_order_id": "098814e2-f46e-4be6-9730-87857812e82a"}	2026-08-01 06:18:24.468094
412	sale	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-08-01 06:18:28.175551	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "0e13ac15-3459-4f35-b2e4-51d47fd7c036", "sales_order_id": "098814e2-f46e-4be6-9730-87857812e82a", "sales_order_number": "ORD-20260801061824"}	2026-08-01 06:18:28.175551
413	sale	sales_order	\N	22	\N	1	\N	\N	\N	1.000	3	\N	\N	2026-08-01 06:18:28.175551	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "52151a2d-fbef-4eb6-bc80-e113d10aa481", "sales_order_id": "098814e2-f46e-4be6-9730-87857812e82a", "sales_order_number": "ORD-20260801061824"}	2026-08-01 06:18:28.175551
414	sale	sales_order	\N	8	\N	1	\N	\N	\N	1.000	1	\N	\N	2026-08-01 06:18:28.175551	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "83e7f800-e9c6-4962-93e4-e1acc13d0aa4", "sales_order_id": "098814e2-f46e-4be6-9730-87857812e82a", "sales_order_number": "ORD-20260801061824"}	2026-08-01 06:18:28.175551
415	allocation	sales_order	\N	35	\N	1	\N	\N	\N	4.000	10	\N	\N	2026-08-01 06:26:48.168901	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "c36ad0d0-1798-47df-be36-55e4de389deb", "sales_order_id": "336b87c3-8860-4889-953b-95145ea1aeab"}	2026-08-01 06:26:48.168901
416	sale	sales_order	\N	35	\N	1	\N	\N	\N	4.000	10	\N	\N	2026-08-01 06:26:52.332477	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "c36ad0d0-1798-47df-be36-55e4de389deb", "sales_order_id": "336b87c3-8860-4889-953b-95145ea1aeab", "sales_order_number": "ORD-20260801062648"}	2026-08-01 06:26:52.332477
417	allocation	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-08-01 06:47:31.594065	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "3e66548b-a5f2-44f3-96b8-c909336a72df", "sales_order_id": "a7dcdda8-9253-4201-a387-a0276ab91c39"}	2026-08-01 06:47:31.594065
418	allocation	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-08-01 06:47:31.602394	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "95d04ad6-96ac-476b-b699-322692aaccba", "sales_order_id": "a7dcdda8-9253-4201-a387-a0276ab91c39"}	2026-08-01 06:47:31.602394
419	sale	sales_order	\N	34	\N	1	\N	\N	\N	1.000	11	\N	\N	2026-08-01 06:47:39.259653	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "3e66548b-a5f2-44f3-96b8-c909336a72df", "sales_order_id": "a7dcdda8-9253-4201-a387-a0276ab91c39", "sales_order_number": "ORD-20260801064731"}	2026-08-01 06:47:39.259653
420	sale	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-08-01 06:47:39.259653	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "95d04ad6-96ac-476b-b699-322692aaccba", "sales_order_id": "a7dcdda8-9253-4201-a387-a0276ab91c39", "sales_order_number": "ORD-20260801064731"}	2026-08-01 06:47:39.259653
421	allocation	sales_order	\N	34	\N	1	\N	\N	\N	10.000	11	\N	\N	2026-08-03 05:31:22.824918	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "3d811adf-1cb9-4323-923d-d7d8393c3ef9", "sales_order_id": "50388992-d30f-4ee7-929d-ee180ede36a7"}	2026-08-03 05:31:22.824918
422	allocation	sales_order	\N	27	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-08-03 05:31:22.834354	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "569ac606-8d30-4e2f-a030-86fb87137d14", "sales_order_id": "50388992-d30f-4ee7-929d-ee180ede36a7"}	2026-08-03 05:31:22.834354
423	allocation	sales_order	\N	23	\N	1	\N	\N	\N	5.000	3	\N	\N	2026-08-03 05:31:22.837677	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "51eedf75-ca64-41ec-b3b2-b52d3c97a275", "sales_order_id": "50388992-d30f-4ee7-929d-ee180ede36a7"}	2026-08-03 05:31:22.837677
460	transfer_in	transfer_request	16	9	\N	4	1	9	2	48.000	5	\N	\N	2026-08-11 07:52:33.453573	2	completed	\N	\N	{"transfer_number": "TRF-COKE-00458"}	2026-08-11 07:52:33.453573
424	sale	sales_order	\N	34	\N	1	\N	\N	\N	10.000	11	\N	\N	2026-08-03 05:31:30.355697	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "3d811adf-1cb9-4323-923d-d7d8393c3ef9", "sales_order_id": "50388992-d30f-4ee7-929d-ee180ede36a7", "sales_order_number": "ORD-20260803053122"}	2026-08-03 05:31:30.355697
425	sale	sales_order	\N	27	\N	1	\N	\N	\N	1.000	8	\N	\N	2026-08-03 05:31:30.355697	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "569ac606-8d30-4e2f-a030-86fb87137d14", "sales_order_id": "50388992-d30f-4ee7-929d-ee180ede36a7", "sales_order_number": "ORD-20260803053122"}	2026-08-03 05:31:30.355697
426	sale	sales_order	\N	23	\N	1	\N	\N	\N	5.000	3	\N	\N	2026-08-03 05:31:30.355697	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "51eedf75-ca64-41ec-b3b2-b52d3c97a275", "sales_order_id": "50388992-d30f-4ee7-929d-ee180ede36a7", "sales_order_number": "ORD-20260803053122"}	2026-08-03 05:31:30.355697
427	allocation	sales_order	\N	6	\N	8	\N	\N	\N	1.000	10	\N	\N	2026-08-03 09:55:00.664779	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "c80bc1c5-4a36-42cc-9bce-557ded221714", "sales_order_id": "ffc609d0-0015-4ced-9d82-025d4f8a7ca3"}	2026-08-03 09:55:00.664779
428	allocation	sales_order	\N	6	\N	8	\N	\N	\N	1.000	10	\N	\N	2026-08-03 09:55:00.673899	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "64cc7612-6f65-4b39-938d-626af1f90a2a", "sales_order_id": "ffc609d0-0015-4ced-9d82-025d4f8a7ca3"}	2026-08-03 09:55:00.673899
429	allocation	sales_order	\N	6	\N	8	\N	\N	\N	1.000	10	\N	\N	2026-08-03 09:55:00.676925	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "412a369a-247d-4617-bdf6-5a9bbf024864", "sales_order_id": "ffc609d0-0015-4ced-9d82-025d4f8a7ca3"}	2026-08-03 09:55:00.676925
430	allocation	sales_order	\N	8	\N	8	\N	\N	\N	1.000	1	\N	\N	2026-08-03 09:55:00.679692	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "62d5d880-cd15-4b6b-bc15-e9e0cfc86cd9", "sales_order_id": "ffc609d0-0015-4ced-9d82-025d4f8a7ca3"}	2026-08-03 09:55:00.679692
431	sale	sales_order	\N	6	\N	8	\N	\N	\N	1.000	10	\N	\N	2026-08-03 09:55:10.831067	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "c80bc1c5-4a36-42cc-9bce-557ded221714", "sales_order_id": "ffc609d0-0015-4ced-9d82-025d4f8a7ca3", "sales_order_number": "ORD-20260803095500"}	2026-08-03 09:55:10.831067
432	sale	sales_order	\N	6	\N	8	\N	\N	\N	1.000	10	\N	\N	2026-08-03 09:55:10.831067	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "64cc7612-6f65-4b39-938d-626af1f90a2a", "sales_order_id": "ffc609d0-0015-4ced-9d82-025d4f8a7ca3", "sales_order_number": "ORD-20260803095500"}	2026-08-03 09:55:10.831067
433	sale	sales_order	\N	6	\N	8	\N	\N	\N	1.000	10	\N	\N	2026-08-03 09:55:10.831067	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "412a369a-247d-4617-bdf6-5a9bbf024864", "sales_order_id": "ffc609d0-0015-4ced-9d82-025d4f8a7ca3", "sales_order_number": "ORD-20260803095500"}	2026-08-03 09:55:10.831067
434	sale	sales_order	\N	8	\N	8	\N	\N	\N	1.000	1	\N	\N	2026-08-03 09:55:10.831067	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "62d5d880-cd15-4b6b-bc15-e9e0cfc86cd9", "sales_order_id": "ffc609d0-0015-4ced-9d82-025d4f8a7ca3", "sales_order_number": "ORD-20260803095500"}	2026-08-03 09:55:10.831067
435	allocation	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-08-03 10:34:21.527587	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "10be4e07-ecfb-4ede-960e-cfbf4da1076d", "sales_order_id": "8fd7a740-50df-4b24-b468-6eabcd6ec92d"}	2026-08-03 10:34:21.527587
436	sale	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-08-03 10:34:25.560444	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "10be4e07-ecfb-4ede-960e-cfbf4da1076d", "sales_order_id": "8fd7a740-50df-4b24-b468-6eabcd6ec92d", "sales_order_number": "ORD-20260803103421"}	2026-08-03 10:34:25.560444
437	allocation	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-08-03 10:36:05.372569	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "f1b26e62-b6ad-4641-aa17-525ee9454a04", "sales_order_id": "8387e582-d524-4c27-8c35-24ec3f9e8d60"}	2026-08-03 10:36:05.372569
438	sale	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-08-03 10:36:09.241354	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "f1b26e62-b6ad-4641-aa17-525ee9454a04", "sales_order_id": "8387e582-d524-4c27-8c35-24ec3f9e8d60", "sales_order_number": "ORD-20260803103605"}	2026-08-03 10:36:09.241354
439	allocation	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-08-03 10:36:47.663662	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "a23dc953-3288-4ab6-9f49-069787816dd3", "sales_order_id": "ce52051e-4a6e-440e-909a-86dd0703926d"}	2026-08-03 10:36:47.663662
440	allocation	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-08-03 10:36:47.666995	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "e19087e1-0020-4e95-9a91-c178b4d71018", "sales_order_id": "ce52051e-4a6e-440e-909a-86dd0703926d"}	2026-08-03 10:36:47.666995
441	sale	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-08-03 10:36:52.148215	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "a23dc953-3288-4ab6-9f49-069787816dd3", "sales_order_id": "ce52051e-4a6e-440e-909a-86dd0703926d", "sales_order_number": "ORD-20260803103647"}	2026-08-03 10:36:52.148215
442	sale	sales_order	\N	33	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-08-03 10:36:52.148215	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "e19087e1-0020-4e95-9a91-c178b4d71018", "sales_order_id": "ce52051e-4a6e-440e-909a-86dd0703926d", "sales_order_number": "ORD-20260803103647"}	2026-08-03 10:36:52.148215
443	allocation	sales_order	\N	8	\N	8	\N	\N	\N	1.000	1	\N	\N	2026-08-03 11:39:23.341481	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "c5c3d50f-2359-4dab-b2f1-bbf28de60789", "sales_order_id": "480b985a-df97-4c2b-95c3-bef7d77c0e72"}	2026-08-03 11:39:23.341481
444	sale	sales_order	\N	8	\N	8	\N	\N	\N	1.000	1	\N	\N	2026-08-03 11:39:31.856418	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "c5c3d50f-2359-4dab-b2f1-bbf28de60789", "sales_order_id": "480b985a-df97-4c2b-95c3-bef7d77c0e72", "sales_order_number": "ORD-20260803113923"}	2026-08-03 11:39:31.856418
445	allocation	sales_order	\N	5	\N	8	\N	\N	\N	1.000	10	\N	\N	2026-08-03 11:41:39.523518	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "75632adc-5812-468e-b6e3-86845884d7ba", "sales_order_id": "332d8ea2-6299-45fc-b833-ad082299dd72"}	2026-08-03 11:41:39.523518
446	sale	sales_order	\N	5	\N	8	\N	\N	\N	1.000	10	\N	\N	2026-08-03 11:41:43.488103	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "75632adc-5812-468e-b6e3-86845884d7ba", "sales_order_id": "332d8ea2-6299-45fc-b833-ad082299dd72", "sales_order_number": "ORD-20260803114139"}	2026-08-03 11:41:43.488103
447	transfer_out	transfer_request	11	7	\N	1	17	7	12	20.000	10	\N	\N	2026-08-11 07:20:20.748627	1	shipped	\N	\N	{"transfer_number": "TRF-94CD-E3D7"}	2026-08-11 07:20:20.748627
448	transfer_out	transfer_request	11	40	\N	1	17	7	12	2.000	4	\N	\N	2026-08-11 07:20:20.748627	1	shipped	\N	\N	{"transfer_number": "TRF-94CD-E3D7"}	2026-08-11 07:20:20.748627
449	transfer_in	transfer_request	11	7	\N	1	17	7	12	20.000	10	\N	\N	2026-08-11 07:20:32.513374	1	completed	\N	\N	{"transfer_number": "TRF-94CD-E3D7"}	2026-08-11 07:20:32.513374
450	transfer_in	transfer_request	11	40	\N	1	17	7	12	2.000	4	\N	\N	2026-08-11 07:20:32.513374	1	completed	\N	\N	{"transfer_number": "TRF-94CD-E3D7"}	2026-08-11 07:20:32.513374
451	transfer_out	transfer_request	12	5	\N	1	17	1	12	9.000	10	\N	\N	2026-08-11 07:26:24.241143	1	shipped	\N	\N	{"transfer_number": "TRF-934C-EF13"}	2026-08-11 07:26:24.241143
452	transfer_in	transfer_request	12	5	\N	1	17	1	12	9.000	10	\N	\N	2026-08-11 07:26:28.923699	1	completed	\N	\N	{"transfer_number": "TRF-934C-EF13"}	2026-08-11 07:26:28.923699
453	transfer_out	transfer_request	13	11	\N	4	17	9	12	2.000	5	\N	\N	2026-08-11 07:28:58.022557	1	shipped	\N	\N	{"transfer_number": "TRF-6EF8-42CC"}	2026-08-11 07:28:58.022557
454	transfer_in	transfer_request	13	11	\N	4	17	9	12	2.000	5	\N	\N	2026-08-11 07:29:04.819316	1	completed	\N	\N	{"transfer_number": "TRF-6EF8-42CC"}	2026-08-11 07:29:04.819316
461	transfer_out	transfer_request	17	9	\N	4	1	9	2	48.000	5	\N	\N	2026-08-11 07:55:47.281033	1	shipped	\N	\N	{"transfer_number": "TRF-COKE-004512"}	2026-08-11 07:55:47.281033
465	transfer_out	transfer_request	20	9	\N	4	1	9	2	48.000	5	\N	\N	2026-08-11 08:20:17.088309	1	shipped	\N	\N	{"transfer_number": "TRF-COKE-004518"}	2026-08-11 08:20:17.088309
462	transfer_in	transfer_request	17	9	\N	4	1	9	2	48.000	5	\N	\N	2026-08-11 07:56:01.02998	2	completed	\N	\N	{"transfer_number": "TRF-COKE-004512"}	2026-08-11 07:56:01.02998
463	transfer_out	transfer_request	19	9	\N	1	17	2	12	72.000	5	\N	\N	2026-08-11 08:11:26.183267	1	shipped	\N	\N	{"transfer_number": "TRF-2017-F080"}	2026-08-11 08:11:26.183267
464	transfer_in	transfer_request	19	9	\N	1	17	2	12	72.000	5	\N	\N	2026-08-11 08:11:33.406704	1	completed	\N	\N	{"transfer_number": "TRF-2017-F080"}	2026-08-11 08:11:33.406704
466	transfer_out	transfer_request	21	9	\N	4	1	9	2	48.000	5	\N	\N	2026-08-11 09:16:51.275576	1	shipped	\N	\N	{"transfer_number": "TRF-COKE-004518"}	2026-08-11 09:16:51.275576
467	transfer_in	transfer_request	21	9	\N	4	1	9	2	48.000	5	\N	\N	2026-08-11 09:17:02.166042	2	completed	\N	\N	{"transfer_number": "TRF-COKE-004518"}	2026-08-11 09:17:02.166042
468	allocation	sales_order	\N	35	\N	1	\N	\N	\N	2.000	10	\N	\N	2026-08-11 11:24:31.135266	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "e48eb8d1-cc07-40a3-913a-c324e9406fad", "sales_order_id": "0ad0c821-b76b-4f0c-b149-7edf6df71da9"}	2026-08-11 11:24:31.135266
469	sale	sales_order	\N	35	\N	1	\N	\N	\N	2.000	10	\N	\N	2026-08-11 11:24:35.342384	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "e48eb8d1-cc07-40a3-913a-c324e9406fad", "sales_order_id": "0ad0c821-b76b-4f0c-b149-7edf6df71da9", "sales_order_number": "ORD-20260811112431"}	2026-08-11 11:24:35.342384
470	allocation	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-08-11 11:46:45.418351	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "cdcfc1f5-3107-49b8-b323-f2047c76a517", "sales_order_id": "1f34855a-1ff6-4e77-a567-65232d4a4504"}	2026-08-11 11:46:45.418351
471	allocation	sales_order	\N	33	\N	1	\N	\N	\N	2.000	10	\N	\N	2026-08-11 11:46:45.427169	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "6d5f3219-38df-47c5-98d9-25568c1f0ec3", "sales_order_id": "1f34855a-1ff6-4e77-a567-65232d4a4504"}	2026-08-11 11:46:45.427169
472	allocation	sales_order	\N	26	\N	1	\N	\N	\N	3.000	8	\N	\N	2026-08-11 11:46:45.429889	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "6cc11e13-088d-4701-8847-0b5417bf8a98", "sales_order_id": "1f34855a-1ff6-4e77-a567-65232d4a4504"}	2026-08-11 11:46:45.429889
473	allocation	sales_order	\N	7	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-08-11 11:46:45.432996	\N	completed	\N	\N	{"order_status": "pending", "order_line_id": "9292b6e4-3ff7-4b23-9192-ca6d15150d9b", "sales_order_id": "1f34855a-1ff6-4e77-a567-65232d4a4504"}	2026-08-11 11:46:45.432996
474	sale	sales_order	\N	35	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-08-11 11:46:54.854467	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "cdcfc1f5-3107-49b8-b323-f2047c76a517", "sales_order_id": "1f34855a-1ff6-4e77-a567-65232d4a4504", "sales_order_number": "ORD-20260811114645"}	2026-08-11 11:46:54.854467
475	sale	sales_order	\N	33	\N	1	\N	\N	\N	2.000	10	\N	\N	2026-08-11 11:46:54.854467	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "6d5f3219-38df-47c5-98d9-25568c1f0ec3", "sales_order_id": "1f34855a-1ff6-4e77-a567-65232d4a4504", "sales_order_number": "ORD-20260811114645"}	2026-08-11 11:46:54.854467
476	sale	sales_order	\N	26	\N	1	\N	\N	\N	3.000	8	\N	\N	2026-08-11 11:46:54.854467	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "6cc11e13-088d-4701-8847-0b5417bf8a98", "sales_order_id": "1f34855a-1ff6-4e77-a567-65232d4a4504", "sales_order_number": "ORD-20260811114645"}	2026-08-11 11:46:54.854467
477	sale	sales_order	\N	7	\N	1	\N	\N	\N	1.000	10	\N	\N	2026-08-11 11:46:54.854467	\N	completed	\N	\N	{"order_status": "fulfilled", "order_line_id": "9292b6e4-3ff7-4b23-9192-ca6d15150d9b", "sales_order_id": "1f34855a-1ff6-4e77-a567-65232d4a4504", "sales_order_number": "ORD-20260811114645"}	2026-08-11 11:46:54.854467
\.


--
-- Data for Name: stock_reservations; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.stock_reservations (id, reservation_number, product_id, product_variant_id, store_id, reference_type, reference_id, quantity_reserved, reserved_at, expires_at, status, reserved_by, notes, metadata, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: storage_locations; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.storage_locations (id, store_id, code, name, location_type, parent_location_id, is_active, metadata, created_at) FROM stdin;
1	1	RYD-DAIRY	Dairy Section	retail_floor	\N	t	{}	2026-07-18 07:58:31.573111
2	1	RYD-BEVERAGE	Beverage Section	retail_floor	\N	t	{}	2026-07-18 07:58:31.573111
3	1	RYD-FOOD	Food & Groceries	retail_floor	\N	t	{}	2026-07-18 07:58:31.573111
4	1	RYD-FROZEN	Frozen Section	retail_floor	\N	t	{}	2026-07-18 07:58:31.573111
5	1	RYD-PERSONAL	Personal Care	retail_floor	\N	t	{}	2026-07-18 07:58:31.573111
6	1	RYD-HOUSEHOLD	Household Products	retail_floor	\N	t	{}	2026-07-18 07:58:31.573111
7	1	RYD-BACK	Back Storage	backroom	\N	t	{}	2026-07-18 07:58:31.573111
8	4	WH-ZONE-A	Zone A - Dry Goods	warehouse_zone	\N	t	{}	2026-07-18 07:58:31.573111
9	4	WH-ZONE-B	Zone B - Beverages	warehouse_zone	\N	t	{}	2026-07-18 07:58:31.573111
10	4	WH-ZONE-C	Zone C - Cold Storage	warehouse_zone	\N	t	{}	2026-07-18 07:58:31.573111
11	4	WH-ZONE-D	Zone D - Frozen	warehouse_zone	\N	t	{}	2026-07-18 07:58:31.573111
12	17	Nast-002	Nastecsol	shelf	\N	t	\N	2026-07-30 11:45:50.962414
13	10	A1 shelf	zone 2	shelf	\N	t	\N	2026-07-30 11:50:45.962963
\.


--
-- Data for Name: stores; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.stores (id, organization_id, parent_store_id, name, code, store_type, is_warehouse, is_pos_enabled, timezone, is_active, metadata, created_at, updated_at) FROM stdin;
1	1	\N	Qitaf al Ayela	RYD-001	retail	f	t	Asia/Riyadh	t	{"city": "Tabuk", "phone": "+966-11-1234567", "address": "Saudia/Tabuk", "manager": "NasaR"}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
2	1	\N	Qitaf al Qadsyia	JED-001	retail	f	t	Asia/Riyadh	t	{"city": "Tabuk", "phone": "+966-11-1234567", "address": "Saudia/Tabuk", "manager": "NasaR"}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
3	1	\N	Qitaf al Tamaouz	DMM-001	retail	f	t	Asia/Riyadh	t	{"city": "Dammam", "phone": "+966-13-3456789", "address": "King Saud Road, Dammam", "manager": "Omar Al-Otaibi"}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
4	1	\N	Qitaf Warehouse	WH-RYD-001	warehouse	t	t	Asia/Riyadh	t	{"city": "Tabuk", "phone": "+966-11-9876543", "address": "Industrial Area, Tabuk", "manager": "Hassan Al-Mutairi"}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
5	1	\N	Wholesale Center Riyadh	WHSL-RYD-001	wholesale	f	t	Asia/Riyadh	t	{"city": "Tabuk", "phone": "+966-11-9876543", "address": "Industrial Area, Tabuk", "manager": "Hassan Al-Mutairi"}	2026-07-18 07:58:31.573111	2026-07-18 07:58:31.573111
8	1	\N	NasaR Cafe & Restaurant	REST-001	restaurant	f	t	Asia/Riyadh	t	{"city": "Tabuk", "phone": "+966-14-1234567", "address": "King Fahd Road, Tabuk", "manager": "NasaR", "seating_capacity": 50}	2026-07-18 08:03:45.815204	2026-07-18 08:03:45.815204
9	1	\N	test	3743	retail	f	t	Asia/Karachi	t	{"gln": "", "city": "Islamabad", "state": "Pakistan", "address": "Ghouri town 5-A", "location": "Tabuk Saudi Arabia", "address_2": "", "address_3": "", "post_code": "", "tax_office": "", "vat_number": "", "price_lists": [{"id": 3, "code": "PROMO_SAR", "name": "Promotional Price List", "metadata": {}, "valid_to": "2026-08-17", "is_active": true, "created_at": "2026-07-18T07:58:31.573111", "is_default": false, "updated_at": "2026-07-18T07:59:03.504911", "valid_from": "2026-07-18", "currency_code": "SAR", "price_list_type": "promotional"}, {"id": 1, "code": "RETAIL_SAR", "name": "Retail Price List", "metadata": {}, "valid_to": null, "is_active": true, "created_at": "2026-07-18T07:58:31.573111", "is_default": true, "updated_at": "2026-07-18T07:58:31.573111", "valid_from": "2024-01-01", "currency_code": "SAR", "price_list_type": "retail"}, {"id": 2, "code": "WHOLESALE_SAR", "name": "Wholesale Price List", "metadata": {}, "valid_to": null, "is_active": true, "created_at": "2026-07-18T07:58:31.573111", "is_default": false, "updated_at": "2026-07-18T07:58:31.573111", "valid_from": "2024-01-01", "currency_code": "SAR", "price_list_type": "wholesale"}], "street_po_box": "", "payment_options": [{"id": "cash", "icon": "fa-money-bill", "name": "Cash", "enabled": true}, {"id": "credit_card", "icon": "fa-credit-card", "name": "Credit Card", "enabled": true}, {"id": "mobile_wallet", "icon": "fa-mobile-alt", "name": "Mobile Wallet", "enabled": true}, {"id": "bank_transfer", "icon": "fa-university", "name": "Bank Transfer", "enabled": true}]}	2026-07-21 09:51:33.50059	2026-07-21 09:51:33.50059
10	1	\N	Nastecsol	Nast-001	cafe	f	t	Asia/Karachi	t	{"gln": "", "city": "Islamabad", "state": "Pakistan", "address": "Street 17 D block", "location": "PWD", "address_2": "", "address_3": "", "post_code": "", "tax_office": "", "vat_number": "", "price_lists": [{"id": 3, "code": "PROMO_SAR", "name": "Promotional Price List", "metadata": {}, "valid_to": "2026-08-17", "is_active": true, "created_at": "2026-07-18T07:58:31.573111", "is_default": false, "updated_at": "2026-07-18T07:59:03.504911", "valid_from": "2026-07-18", "currency_code": "SAR", "price_list_type": "promotional"}, {"id": 1, "code": "RETAIL_SAR", "name": "Retail Price List", "metadata": {}, "valid_to": null, "is_active": true, "created_at": "2026-07-18T07:58:31.573111", "is_default": true, "updated_at": "2026-07-18T07:58:31.573111", "valid_from": "2024-01-01", "currency_code": "SAR", "price_list_type": "retail"}, {"id": 2, "code": "WHOLESALE_SAR", "name": "Wholesale Price List", "metadata": {}, "valid_to": null, "is_active": true, "created_at": "2026-07-18T07:58:31.573111", "is_default": false, "updated_at": "2026-07-18T07:58:31.573111", "valid_from": "2024-01-01", "currency_code": "SAR", "price_list_type": "wholesale"}], "street_po_box": "", "payment_options": [{"id": "cash", "icon": "fa-money-bill", "name": "Cash", "enabled": true}, {"id": "debit_card", "icon": "fa-credit-card", "name": "Debit Card", "enabled": true}]}	2026-07-30 05:31:42.467646	2026-07-30 05:31:42.467646
17	1	\N	nastecsol pos	Nast-002	retail	f	t	Asia/Karachi	t	{"gln": "", "city": "Islamabad", "state": "Pakistan", "address": "ghouri twon phase 5a house no m17 street 1a islamabad", "location": "PWD", "address_2": "", "address_3": "", "post_code": "", "tax_office": "", "vat_number": "", "price_lists": [{"id": 1, "code": "RETAIL_SAR", "name": "Retail Price List", "metadata": {}, "valid_to": null, "is_active": true, "created_at": "2026-07-18T07:58:31.573111", "is_default": true, "updated_at": "2026-07-18T07:58:31.573111", "valid_from": "2024-01-01", "currency_code": "SAR", "price_list_type": "retail"}], "street_po_box": "", "payment_options": [{"id": "cash", "icon": "fa-money-bill", "name": "Cash", "enabled": true}]}	2026-07-30 11:42:55.022544	2026-07-30 11:42:55.022544
\.


--
-- Data for Name: submenu_permissions; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.submenu_permissions (id, submenu_id, permission_id, metadata) FROM stdin;
1	1	1	{}
2	1	2	{}
3	2	1	{}
4	3	1	{}
5	4	1	{}
6	5	4	{}
7	6	5	{}
8	7	7	{}
9	8	8	{}
10	9	9	{}
11	10	11	{}
12	11	12	{}
13	12	11	{}
14	13	15	{}
15	14	16	{}
16	15	15	{}
17	15	22	{}
18	16	25	{}
19	17	26	{}
20	18	28	{}
21	19	25	{}
22	20	26	{}
23	21	29	{}
24	22	31	{}
25	23	32	{}
26	24	29	{}
27	25	30	{}
28	26	35	{}
29	27	35	{}
30	28	36	{}
31	29	37	{}
32	30	40	{}
33	31	40	{}
34	32	39	{}
35	33	39	{}
36	34	41	{}
37	35	41	{}
40	38	44	{}
41	39	44	{}
42	40	46	{}
43	41	47	{}
44	42	47	{}
45	43	46	{}
46	44	47	{}
47	45	46	{}
48	46	47	{}
49	47	49	{}
50	48	49	{}
51	49	51	{}
52	50	52	{}
53	51	54	{}
54	52	55	{}
55	53	56	{}
56	54	58	{}
57	55	59	{}
58	56	61	{}
59	57	62	{}
60	58	63	{}
61	59	66	{}
62	60	66	{}
63	61	66	{}
64	62	67	{}
65	63	67	{}
66	64	68	{}
67	65	68	{}
68	66	69	{}
69	67	69	{}
70	68	19	{}
71	69	20	{}
72	70	22	{}
73	71	71	{}
74	72	72	{}
75	73	73	{}
84	96	42	{}
85	97	42	{}
86	98	47	{}
87	99	59	{}
88	34	97	{}
96	97	103	{}
97	97	104	{}
98	97	97	{}
99	104	102	{}
100	105	102	{}
101	106	90	{}
102	107	89	{}
103	108	89	{}
104	109	91	{}
105	110	92	{}
106	111	89	{}
107	112	89	{}
108	114	20	{"scope": "all"}
109	115	59	{}
110	116	101	{}
111	117	101	{}
112	118	92	{}
113	118	34	{}
114	118	101	{}
115	118	102	{}
116	118	33	{}
117	118	122	{}
118	120	123	{}
119	120	124	{}
120	120	125	{}
121	120	126	{}
122	120	127	{}
123	121	128	{}
124	122	129	{}
125	123	130	{}
126	125	131	{}
127	125	132	{}
128	125	133	{}
129	133	134	{}
130	133	135	{}
131	133	136	{}
132	133	137	{}
133	133	138	{}
134	133	139	{}
135	126	141	{}
136	126	142	{}
137	124	143	{}
138	124	144	{}
139	127	145	{}
140	128	146	{}
141	128	147	{}
142	129	148	{}
143	130	149	{}
144	131	150	{}
145	132	151	{}
\.


--
-- Data for Name: submenus; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.submenus (id, menu_id, parent_submenu_id, name, code, route_path, icon, display_order, is_active, metadata, created_at, updated_at) FROM stdin;
1	1	\N	Admin Dashboard	admin_dashboard	/dashboard/admin	layout	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
2	1	\N	Store Dashboard	store_dashboard	/dashboard/store	store	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
3	2	\N	Sales Analytics	sales_analytics	/dashboard/analytics/sales	trending-up	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
4	2	\N	Inventory Analytics	inventory_analytics	/dashboard/analytics/inventory	package	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
5	3	\N	Tenant List	tenant_list	/admin/tenants/list	list	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
6	3	\N	Add Tenant	add_tenant	/admin/tenants/new	plus	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
7	3	\N	Tenant Configuration	tenant_config	/admin/tenants/config	settings	3	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
8	4	\N	Organization List	org_list	/admin/organizations/list	list	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
9	4	\N	Add Organization	add_org	/admin/organizations/new	plus	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
10	5	\N	User List	user_list	/admin/users/list	list	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
11	5	\N	Add User	add_user	/admin/users/new	user-plus	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
12	5	\N	User Activity	user_activity	/admin/users/activity	activity	3	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
13	6	\N	Role List	role_list	/admin/roles/list	shield	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
14	6	\N	Add Role	add_role	/admin/roles/new	plus	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
15	6	\N	Permission Matrix	permission_matrix	/admin/roles/permissions	grid	3	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
16	7	\N	Store List	store_list	/stores/list	list	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
17	7	\N	Add Store	add_store	/stores/new	plus	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
18	7	\N	Store Configuration	store_config	/stores/config	settings	3	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
19	8	\N	Location List	location_list	/stores/locations/list	list	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
20	8	\N	Add Location	add_location	/stores/locations/new	plus	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
21	9	\N	Transaction List	transaction_list	/pos/transactions/list	list	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
22	9	\N	Process Sale	process_sale	/pos/transactions/new	shopping-cart	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
23	9	\N	Void Transaction	void_transaction	/pos/transactions/void	x-circle	3	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
24	10	\N	Terminal List	terminal_list	/pos/terminals/list	list	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
25	10	\N	Add Terminal	add_terminal	/pos/terminals/new	plus	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
26	11	\N	Daily Sales Report	daily_sales	/pos/reports/daily	calendar	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
27	11	\N	Cashier Performance	cashier_performance	/pos/reports/cashier	award	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
28	12	\N	Cashier List	cashier_list	/cashiers/list	list	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
29	12	\N	Add Cashier	add_cashier	/cashiers/new	user-plus	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
30	13	\N	Active Sessions	active_sessions	/cashiers/sessions/active	clock	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
31	13	\N	Session History	session_history	/cashiers/sessions/history	history	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
32	13	\N	Open Session	open_session	/cashiers/sessions/open	unlock	3	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
33	13	\N	Close Session	close_session	/cashiers/sessions/close	lock	4	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
34	14	\N	Stock Levels	stock_levels	/inventory/overview/levels	package	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
35	14	\N	Low Stock Alert	low_stock	/inventory/overview/low-stock	alert-triangle	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
38	16	\N	Stock Count List	stock_count_list	/inventory/counts/list	list	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
39	16	\N	Create Count	create_count	/inventory/counts/new	plus	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
40	17	\N	Product List	product_list	/products/list	list	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
41	17	\N	Add Product	add_product	/products/new	plus	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
42	17	\N	Product Import	product_import	/products/import	upload	3	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
43	18	\N	Category List	category_list	/products/categories/list	list	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
44	18	\N	Add Category	add_category	/products/categories/new	plus	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
45	19	\N	Brand List	brand_list	/products/brands/list	list	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
46	19	\N	Add Brand	add_brand	/products/brands/new	plus	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
47	20	\N	Price List Management	price_list_mgmt	/products/price-lists/list	list	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
48	20	\N	Add Price List	add_price_list	/products/price-lists/new	plus	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
49	21	\N	Customer List	customer_list	/customers/list	list	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
50	21	\N	Add Customer	add_customer	/customers/new	user-plus	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
51	21	\N	Customer History	customer_history	/customers/history	history	3	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
52	22	\N	Supplier List	supplier_list	/suppliers/list	list	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
53	22	\N	Add Supplier	add_supplier	/suppliers/new	plus	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
54	23	\N	PO List	po_list	/purchase-orders/list	list	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
55	23	\N	Create PO	create_po	/purchase-orders/new	plus	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
56	23	\N	Approve PO	approve_po	/purchase-orders/approve	check-circle	3	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
57	24	\N	SO List	so_list	/sales-orders/list	list	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
58	24	\N	Create SO	create_so	/sales-orders/new	plus	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
59	25	\N	Daily Sales	daily_sales_report	/reports/sales/daily	calendar	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
60	25	\N	Monthly Sales	monthly_sales_report	/reports/sales/monthly	calendar	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
61	25	\N	Product Performance	product_performance	/reports/sales/products	trending-up	3	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
62	26	\N	Purchase Summary	purchase_summary	/reports/purchases/summary	file-text	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
63	26	\N	Supplier Analysis	supplier_analysis	/reports/purchases/suppliers	truck	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
64	27	\N	Stock Valuation	stock_valuation	/reports/inventory/valuation	dollar-sign	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
65	27	\N	Inventory Turnover	inventory_turnover	/reports/inventory/turnover	refresh-cw	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
66	28	\N	Profit & Loss	profit_loss	/reports/financial/pl	trending-up	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
67	28	\N	Discount Analysis	discount_analysis	/reports/financial/discounts	percent	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
68	29	\N	Module List	module_list	/admin/ui-modules/list	list	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
69	29	\N	Menu Management	menu_management	/admin/ui-modules/menus	menu	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
70	29	\N	Permission Management	permission_management	/admin/ui-modules/permissions	lock	3	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
71	30	\N	General Settings	general_settings	/admin/settings/general	settings	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
72	30	\N	Tax Configuration	tax_config	/admin/settings/tax	file-text	2	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
73	31	\N	View Audit Logs	view_audit_logs	/admin/audit-logs/view	eye	1	t	{}	2026-07-18 07:59:19.830838	2026-07-18 07:59:19.830838
96	14	\N	Movement History	movement_history	/inventory/overview/history	warehouse	3	t	"{\\"color\\":\\"blue\\"}"	2026-07-20 04:52:52.607126	2026-07-20 04:52:52.607126
97	14	\N	Record Movement	record_movement	/inventory/overview/new	shipping	4	t	"{\\"color\\":\\"blue\\"}"	2026-07-20 04:57:40.539995	2026-07-20 04:57:40.539995
98	17	\N	UOM	UOM	/products/uom	menu	4	t	"{\\"color\\":\\"blue\\"}"	2026-07-21 07:45:57.361925	2026-07-21 07:45:57.361925
99	23	\N	Procurement Dashboard	procurement_dashboard	/purchase-orders/procurement	bank	4	t	"{\\"color\\":\\"blue\\"}"	2026-07-28 06:36:55.411003	2026-07-28 06:36:55.411003
104	48	\N	Promotions List	promo_list	promotions-discount/promotion/list	chart	1	t	"{\\"color\\":\\"blue\\"}"	2026-07-31 05:57:33.422627	2026-07-31 05:57:33.422627
105	49	\N	Coupon List	coupon_list	promotions-discount/coupon/list	chart	1	t	"{\\"color\\":\\"blue\\"}"	2026-07-31 05:58:17.410278	2026-07-31 05:58:17.410278
106	50	\N	Recipes	recipes	restaurant/catalog/recipes	adjustments	1	t	"{\\"color\\":\\"blue\\"}"	2026-08-01 07:52:35.348608	2026-08-01 07:52:35.348608
107	50	\N	Menu Items	menu_items	restaurant/catalog/menu	clipboard-list	2	t	"{\\"color\\":\\"blue\\"}"	2026-08-01 07:57:47.364219	2026-08-01 07:57:47.364219
108	50	\N	Menu Categories	menu_categories	restaurant/catalog/menu-categories	database	3	t	"{\\"color\\":\\"blue\\"}"	2026-08-01 08:00:40.686517	2026-08-01 08:00:40.686517
109	51	\N	Table Management	restaurant_tables	restaurant/dining/table	inventory	1	t	"{\\"color\\":\\"blue\\"}"	2026-08-01 08:02:05.03719	2026-08-01 08:02:05.03719
110	51	\N	Active Orders	active_restaurant_orders	restaurant/dining/actve-orders	shipping	2	t	"{\\"color\\":\\"blue\\"}"	2026-08-01 08:04:35.722413	2026-08-01 08:04:35.722413
111	52	\N	Waste Logs	waste_logs	restaurant/kitchen	clock	1	t	"{\\"color\\":\\"blue\\"}"	2026-08-01 08:13:13.627857	2026-08-01 08:13:13.627857
112	50	\N	Modifiers & Groups	menu_modifiers	restaurant/catalog/menu-modifiers	payments	4	t	"{\\"color\\":\\"blue\\"}"	2026-08-03 09:29:02.041552	2026-08-03 09:29:02.041552
114	29	\N	Submenu Management	submenu_management	/admin/ui-modules/submenu	user-icon	4	t	"{\\"color\\":\\"blue\\"}"	2026-08-10 12:10:36.068045	2026-08-10 12:10:36.068045
115	23	\N	Quotation Management	quotation_management	/purchase-orders/quotation	shipping	5	t	"{\\"color\\":\\"blue\\"}"	2026-08-10 12:28:41.710739	2026-08-10 12:28:41.710739
116	48	\N	Add Promotion	add_promo	/promotions-discount/promotion/new	plus	2	t	"{\\"color\\":\\"blue\\"}"	2026-08-11 05:35:19.437241	2026-08-11 05:35:19.437241
117	49	\N	Add Coupon	add-coupon	/promotions-discount/coupon/new	plus	2	t	"{\\"color\\":\\"blue\\"}"	2026-08-11 05:36:39.721861	2026-08-11 05:36:39.721861
118	53	\N	POS Dashboard	pos-dashboard	/pos/dashboard	dashboard	1	t	"{\\"color\\":\\"blue\\"}"	2026-08-11 06:53:36.168393	2026-08-11 06:53:36.168393
120	53	\N	POS Product list	pos-products	/dashboard/items-grid	inventory	3	t	"{\\"color\\":\\"blue\\"}"	2026-08-11 06:57:30.310656	2026-08-11 06:57:30.310656
121	53	\N	Retail POS	retail-pos	/dashboard/items-grid	cart	3	t	"{\\"color\\":\\"blue\\"}"	2026-08-11 06:58:21.079106	2026-08-11 06:58:21.079106
122	53	\N	Wholesale POS	wholesale-pos	/dashboard/items-grid	cart	5	t	"{\\"color\\":\\"blue\\"}"	2026-08-11 06:59:11.689128	2026-08-11 06:59:11.689128
123	53	\N	Restaurant POS	restaurant-pos	/dashboard/restaurant-tables	menu	6	t	"{\\"color\\":\\"blue\\"}"	2026-08-11 07:05:56.108093	2026-08-11 07:05:56.108093
124	53	\N	Stock Levels	posstocklevels	/pos/dashboard	trending	7	t	"{\\"color\\":\\"blue\\"}"	2026-08-11 07:06:47.891152	2026-08-11 07:06:47.891152
125	53	\N	POS Bills	pos-bills	/dashboard/bills	document	2	t	"{\\"color\\":\\"blue\\"}"	2026-08-11 07:07:49.986762	2026-08-11 07:07:49.986762
126	53	\N	Cash Collection	CASH_REGISTER	/dashboard/cashcollection	menu	1	t	"{\\"color\\":\\"blue\\"}"	2026-08-11 07:08:33.691154	2026-08-11 07:08:33.691154
127	54	\N	Customer List	customer_list	/pos/dashboard	admin	1	t	"{\\"color\\":\\"blue\\"}"	2026-08-11 07:09:56.601112	2026-08-11 07:09:56.601112
128	54	\N	Add Customer	add_customer	/pos/dashboard	plus	1	t	"{\\"color\\":\\"blue\\"}"	2026-08-11 07:10:35.618251	2026-08-11 07:10:35.618251
129	55	\N	Open session	open-session	/pos/dashboard	lock-open	1	t	"{\\"color\\":\\"blue\\"}"	2026-08-11 07:12:16.980126	2026-08-11 07:12:16.980126
130	55	\N	Close session	close-session	/pos/dashboard	lock	2	t	"{\\"color\\":\\"blue\\"}"	2026-08-11 07:12:58.007269	2026-08-11 07:12:58.007269
131	55	\N	Active session	active-session	/pos/dashboard	clock	3	t	"{\\"color\\":\\"blue\\"}"	2026-08-11 07:14:02.697433	2026-08-11 07:14:02.697433
132	55	\N	Session History	session-history	/pos/dashboard	report	4	t	"{\\"color\\":\\"blue\\"}"	2026-08-11 07:15:04.420236	2026-08-11 07:15:04.420236
133	53	\N	POS Cart	pos-cart	/dashboard/items-grid	store	10	t	"{\\"color\\":\\"blue\\"}"	2026-08-11 07:27:24.721907	2026-08-11 07:27:24.721907
\.


--
-- Data for Name: suppliers; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.suppliers (id, organization_id, code, name, supplier_type, credit_limit, contact_person, email, phone, address, currency_code, payment_terms, tax_id, is_active, metadata, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: tax_categories; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.tax_categories (id, name, code, tax_rate, is_inclusive, is_active, metadata, created_at) FROM stdin;
1	Standard VAT	VAT_15	15.00	f	t	{}	2026-07-18 07:58:31.573111
2	Zero Rated	VAT_0	0.00	f	t	{}	2026-07-18 07:58:31.573111
3	Exempt	VAT_EXEMPT	0.00	f	t	{}	2026-07-18 07:58:31.573111
\.


--
-- Data for Name: tenants; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.tenants (id, tenant_name, slug, db_conn_str, is_active, settings, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: transfer_request_items; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.transfer_request_items (id, transfer_request_id, product_id, product_variant_id, from_location_id, to_location_id, requested_quantity, approved_quantity, shipped_quantity, received_quantity, uom_id, batch_number, notes, created_at) FROM stdin;
21	21	9	\N	9	2	2.000	2.000	2.000	2.000	5	\N	\N	2026-08-11 09:16:06.980858
\.


--
-- Data for Name: transfer_requests; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.transfer_requests (id, organization_id, transfer_number, from_store_id, to_store_id, status, requested_by, approved_by, shipped_by, received_by, request_date, expected_delivery_date, shipped_at, received_at, notes, metadata, created_at, updated_at) FROM stdin;
21	1	TRF-COKE-004518	4	1	received	\N	1	1	2	2026-08-11 09:16:06.971571	\N	2026-08-11 09:16:51.275576	2026-08-11 09:17:02.166042	Testing standard Coke Carton transfer (2 Cartons = 48 base Cans deduction)	{"history": [{"notes": "Testing standard Coke Carton transfer (2 Cartons = 48 base Cans deduction)", "status": "draft", "changed_at": "2026-08-11T09:16:06.972496+00:00", "user_details": {"id": null, "username": "system"}, "transfer_items_snapshot": []}, {"notes": "Testing standard Coke Carton transfer (2 Cartons = 48 base Cans deduction)", "status": "approved", "changed_at": "2026-08-11T09:16:18.579482+00:00", "user_details": {"id": 1, "username": "admin"}, "transfer_items_snapshot": [{"sku": "COCA-COLA-330ML", "uom": "CTN", "base_uom": "CAN", "product_id": 9, "product_name": "Coca-Cola 330ml Can", "shipped_quantity": 0.000, "received_quantity": 0.000, "inventory_snapshot": {"source_store": {"deducted": 0, "store_name": "Qitaf Warehouse", "after_on_hand": 428.000, "before_on_hand": 428.000}, "destination_store": {"store_name": "Qitaf al Ayela", "after_on_hand": 315.000, "added_received": 0, "before_on_hand": 315.000, "after_in_transit": 48.000, "before_in_transit": 48.000}}, "product_variant_id": null, "requested_quantity": 2.000, "converted_base_quantity": 48.000000000000000}]}, {"notes": "Testing standard Coke Carton transfer (2 Cartons = 48 base Cans deduction)", "status": "shipped", "changed_at": "2026-08-11T09:16:51.275576+00:00", "user_details": {"id": 1, "username": "admin"}, "transfer_items_snapshot": [{"sku": "COCA-COLA-330ML", "uom": "CTN", "base_uom": "CAN", "product_id": 9, "product_name": "Coca-Cola 330ml Can", "shipped_quantity": 2.000, "received_quantity": 0.000, "inventory_snapshot": {"source_store": {"deducted": 48.000000000000000, "store_name": "Qitaf Warehouse", "after_on_hand": 380.000, "before_on_hand": 428.000000000000000}, "destination_store": {"store_name": "Qitaf al Ayela", "after_on_hand": 315.000, "added_received": 0, "before_on_hand": 315.000, "after_in_transit": 96.000, "before_in_transit": 48.000000000000000}}, "product_variant_id": null, "requested_quantity": 2.000, "converted_base_quantity": 48.000000000000000}]}, {"notes": "Testing standard Coke Carton transfer (2 Cartons = 48 base Cans deduction)", "status": "received", "changed_at": "2026-08-11T09:17:02.166042+00:00", "user_details": {"id": 2, "username": "owner"}, "transfer_items_snapshot": [{"sku": "COCA-COLA-330ML", "uom": "CTN", "base_uom": "CAN", "product_id": 9, "product_name": "Coca-Cola 330ml Can", "shipped_quantity": 2.000, "received_quantity": 2.000, "inventory_snapshot": {"source_store": {"deducted": 0, "store_name": "Qitaf Warehouse", "after_on_hand": 380.000, "before_on_hand": 380.000}, "destination_store": {"store_name": "Qitaf al Ayela", "after_on_hand": 363.000, "added_received": 48.000000000000000, "before_on_hand": 315.000000000000000, "after_in_transit": 48.000, "before_in_transit": 96.000000000000000}}, "product_variant_id": null, "requested_quantity": 2.000, "converted_base_quantity": 48.000000000000000}]}]}	2026-08-11 09:16:06.972496	2026-08-11 09:17:02.166042
\.


--
-- Data for Name: ui_settings; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.ui_settings (id, submenu_id, setting_key, setting_value, description, metadata, created_at, updated_at) FROM stdin;
1	10	table_columns	{"columns": ["username", "email", "first_name", "last_name", "is_active", "created_at"], "default_sort": "created_at", "default_order": "desc"}	User list table configuration	{}	2026-07-18 07:59:38.038245	2026-07-18 07:59:38.038245
2	10	pagination	{"enabled": true, "default_page_size": 25, "page_size_options": [10, 25, 50, 100]}	User list pagination settings	{}	2026-07-18 07:59:38.038245	2026-07-18 07:59:38.038245
3	21	table_columns	{"columns": ["transaction_number", "cashier_name", "customer_name", "total_amount", "status", "transaction_date"], "default_sort": "transaction_date", "default_order": "desc"}	POS transaction list configuration	{}	2026-07-18 07:59:38.038245	2026-07-18 07:59:38.038245
4	34	alert_threshold	{"show_alerts": true, "low_stock_threshold": 10}	Inventory alert settings	{}	2026-07-18 07:59:38.038245	2026-07-18 07:59:38.038245
5	40	display_mode	{"view": "grid", "items_per_page": 20}	Product list display settings	{}	2026-07-18 07:59:38.038245	2026-07-18 07:59:38.038245
\.


--
-- Data for Name: units_of_measure; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.units_of_measure (id, code, name, uom_type, decimal_places, is_active, metadata) FROM stdin;
1	PCS	Pieces	quantity	0	t	{}
2	KG	Kilogram	weight	3	t	{}
3	LTR	Liter	volume	3	t	{}
4	BOX	Box	packaging	0	t	{}
5	CTN	Carton	packaging	0	t	{}
6	PKT	Packet	packaging	0	t	{}
7	BTL	Bottle	packaging	0	t	{}
8	CAN	Can	packaging	0	t	{}
10	GM	Gram	weight	0	t	{}
11	ML	Milliliter	volume	0	t	{}
12	DZN	Dozen	quantity	0	t	{}
13	TRAY	Tray	packaging	0	t	{"description": "Plastic or cardboard tray"}
14	BUNDLE	Bundle	packaging	0	t	{"description": "Bundle of items"}
15	PACK	Pack	packaging	0	t	{"description": "Small pack"}
16	SACK	Sack	packaging	0	t	{"description": "Large sack for bulk items"}
17	PALLET	Pallet	packaging	0	t	{"description": "Full pallet"}
18	CASE	Case	packaging	0	t	{"description": "Case or container"}
9	BAG	Bag	packaging	1	t	{"packaging_templates": [{"name": "Template A", "hierarchy": [{"uom_code": "BAG", "uom_name": "Bag", "multiplier": 1, "level_order": 1}]}]}
20	BEG	حقيبة	packaging	1	t	{"packaging_templates": [{"name": "Template A", "hierarchy": [{"uom_code": "BEG", "uom_name": "حقيبة", "multiplier": 1, "level_order": 1}]}]}
\.


--
-- Data for Name: uom_packaging_template_levels; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.uom_packaging_template_levels (id, template_id, level_order, uom_id, multiplier) FROM stdin;
1	1	1	1	1.000000
2	1	2	2	24.000000
3	1	3	3	12.000000
4	2	1	1	1.000000
5	2	2	2	12.000000
6	2	3	3	6.000000
7	3	1	1	1.000000
8	3	2	4	50.000000
9	3	3	5	10.000000
10	4	1	1	1.000000
11	4	2	6	6.000000
12	4	3	7	4.000000
13	5	1	1	1.000000
14	5	2	8	100.000000
15	5	3	9	10.000000
\.


--
-- Data for Name: uom_packaging_templates; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.uom_packaging_templates (id, organization_id, name, code, is_active, created_at, updated_at) FROM stdin;
1	1	Beverage Standard Pattern	1-24-12	t	2026-07-18 07:59:38.038245	2026-07-18 07:59:38.038245
2	1	Snack Box Pattern	1-12-6	t	2026-07-18 07:59:38.038245	2026-07-18 07:59:38.038245
3	1	Warehouse Bulk Pattern	1-50-10	t	2026-07-18 07:59:38.038245	2026-07-18 07:59:38.038245
4	1	Retail Small Pattern	1-6-4	t	2026-07-18 07:59:38.038245	2026-07-18 07:59:38.038245
5	1	Pharma Packaging Pattern	1-100-10	t	2026-07-18 07:59:38.038245	2026-07-18 07:59:38.038245
\.


--
-- Data for Name: user_roles; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.user_roles (id, user_id, role_id, metadata, assigned_at) FROM stdin;
1	1	1	{}	2026-07-18 07:59:38.038245
2	2	2	{}	2026-07-18 07:59:38.038245
3	3	3	{}	2026-07-18 07:59:38.038245
4	4	4	{}	2026-07-18 07:59:38.038245
5	5	5	{}	2026-07-18 07:59:38.038245
7	12	4	{}	2026-07-23 12:09:38.874528
8	13	4	{}	2026-07-23 12:36:17.810327
27	31	4	{}	2026-07-28 05:44:12.6822
30	44	4	{}	2026-07-30 10:39:33.397817
\.


--
-- Data for Name: user_store_access; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.user_store_access (id, user_id, store_id, is_primary, metadata, granted_at) FROM stdin;
1	1	1	t	{}	2026-07-18 07:59:38.038245
2	1	2	f	{}	2026-07-18 07:59:38.038245
3	1	3	f	{}	2026-07-18 07:59:38.038245
4	2	1	t	{}	2026-07-18 07:59:38.038245
5	2	2	f	{}	2026-07-18 07:59:38.038245
6	2	3	f	{}	2026-07-18 07:59:38.038245
7	3	1	t	{}	2026-07-18 07:59:38.038245
8	4	1	t	{}	2026-07-18 07:59:38.038245
9	5	3	t	{}	2026-07-18 07:59:38.038245
10	5	1	f	{}	2026-07-18 07:59:38.038245
12	4	9	f	null	2026-07-21 09:51:34.500299
13	12	8	t	{}	2026-07-23 12:09:38.878854
14	13	8	t	{}	2026-07-23 12:36:17.813014
40	31	1	t	{}	2026-07-28 05:44:12.686119
43	1	10	f	null	2026-07-30 05:31:43.49255
44	2	10	f	null	2026-07-30 05:31:43.998184
45	4	10	f	null	2026-07-30 05:31:44.503719
46	31	10	f	null	2026-07-30 05:31:45.041616
47	13	10	f	null	2026-07-30 05:31:45.558022
48	44	10	t	{}	2026-07-30 10:39:33.401527
49	44	17	f	null	2026-07-30 11:42:55.628941
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.users (id, organization_id, username, email, password_hash, first_name, last_name, employee_code, is_active, metadata, created_at, updated_at) FROM stdin;
3	1	manager	manager@democorp.com	$2a$10$dcbOmMwFBzWWMJWhSy3iW.HasRcquRCFz5nRmcIMK36V6VCxoIlCC	Store	Manager	EMP003	t	{}	2026-07-18 07:59:38.038245	2026-07-18 07:59:38.038245
4	1	cashier1	cashier1@democorp.com	$2a$10$dcbOmMwFBzWWMJWhSy3iW.HasRcquRCFz5nRmcIMK36V6VCxoIlCC	John	Cashier	EMP004	t	{}	2026-07-18 07:59:38.038245	2026-07-18 07:59:38.038245
5	1	inventory	inventory@democorp.com	$2a$10$dcbOmMwFBzWWMJWhSy3iW.HasRcquRCFz5nRmcIMK36V6VCxoIlCC	Inventory	Manager	EMP005	t	{}	2026-07-18 07:59:38.038245	2026-07-18 07:59:38.038245
1	1	admin	admin@democorp.com	$2a$10$dcbOmMwFBzWWMJWhSy3iW.HasRcquRCFz5nRmcIMK36V6VCxoIlCC	Admin	User	EMP001	t	null	2026-07-18 07:59:38.038245	2026-07-27 06:55:42.500354
12	1	testing	adil	$2a$10$Q5usbEbTPd7rUOHkcEXE6ehIEfjCJqEtG8Z12g7Ec.A9r23ZeZbfG	new	test	\N	t	null	2026-07-23 12:09:38.379403	2026-07-27 08:18:08.315071
13	1	QGL153	adilfarooq3540@gmail.com	$2a$10$mBbd69ZY7R5ZBamlP5w3KeiSZtl6TLamCrlLXHHj/pQwc1f47Mdjq	test1	123	\N	t	null	2026-07-23 12:36:17.258415	2026-07-29 12:33:07.026328
44	1	abuhurrairajaved	hurrairajaved123@gmail.com	$2a$10$0URVU9h71pVbjl7EVJAcYeMF/StdcbyrHbgMyp1oVnsaPgTu5ePLy	ABUHURRAIRA	JAVED	\N	t	{}	2026-07-30 10:39:33.14388	2026-07-30 10:39:33.14388
31	1	cashierr4	cashier@example.com	$2a$10$MJwSbOkJ2ccFPyGo4WuJ.e3/qi.pIRvtfvw8duvsTDG45BlugDJ8y	cashierr	4	\N	t	{}	2026-07-28 05:44:12.180849	2026-07-28 05:44:12.180849
2	1	owner	owner@democorp.com	$2a$10$dcbOmMwFBzWWMJWhSy3iW.HasRcquRCFz5nRmcIMK36V6VCxoIlCC	Business	Owner	EMP002	t	{"id": 84721, "age": 29, "name": "John Doe", "email": "john.doe@example.com", "roles": ["admin", "editor"], "address": {"city": "Springfield", "state": "Illinois", "street": "42 Maple Avenue", "country": "USA", "zipCode": "62704"}, "isActive": true}	2026-07-18 07:59:38.038245	2026-07-28 08:04:31.492071
\.


--
-- Data for Name: waste_logs; Type: TABLE DATA; Schema: public; Owner: nembus_admin_user
--

COPY public.waste_logs (id, store_id, product_id, menu_item_id, recipe_id, waste_source, quantity, uom_id, unit_cost, total_cost, reason, logged_by, order_id, wasted_at, metadata, created_at) FROM stdin;
\.


--
-- Name: audit_logs_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.audit_logs_id_seq', 1, false);


--
-- Name: brands_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.brands_id_seq', 26, true);


--
-- Name: cart_activity_log_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.cart_activity_log_id_seq', 67, true);


--
-- Name: cashier_sessions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.cashier_sessions_id_seq', 39, true);


--
-- Name: cashiers_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.cashiers_id_seq', 6, true);


--
-- Name: combo_bundle_items_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.combo_bundle_items_id_seq', 1, false);


--
-- Name: combo_bundles_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.combo_bundles_id_seq', 1, false);


--
-- Name: customers_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.customers_id_seq', 15, true);


--
-- Name: discount_analytics_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.discount_analytics_id_seq', 1, false);


--
-- Name: inventory_analytics_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.inventory_analytics_id_seq', 1, false);


--
-- Name: inventory_stock_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.inventory_stock_id_seq', 224, true);


--
-- Name: invoice_status_history_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.invoice_status_history_id_seq', 1, false);


--
-- Name: kiosk_sessions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.kiosk_sessions_id_seq', 1, false);


--
-- Name: loyalty_redemption_rules_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.loyalty_redemption_rules_id_seq', 1, false);


--
-- Name: menu_categories_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.menu_categories_id_seq', 15, true);


--
-- Name: menu_item_availability_schedules_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.menu_item_availability_schedules_id_seq', 1, false);


--
-- Name: menu_item_modifiers_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.menu_item_modifiers_id_seq', 1, false);


--
-- Name: menu_items_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.menu_items_id_seq', 10, true);


--
-- Name: menu_modifier_groups_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.menu_modifier_groups_id_seq', 1, false);


--
-- Name: menu_permissions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.menu_permissions_id_seq', 222, true);


--
-- Name: menus_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.menus_id_seq', 55, true);


--
-- Name: module_permissions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.module_permissions_id_seq', 135, true);


--
-- Name: modules_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.modules_id_seq', 22, true);


--
-- Name: order_status_history_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.order_status_history_id_seq', 1, false);


--
-- Name: organizations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.organizations_id_seq', 1, true);


--
-- Name: permissions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.permissions_id_seq', 151, true);


--
-- Name: pos_payments_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.pos_payments_id_seq', 62, true);


--
-- Name: pos_terminals_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.pos_terminals_id_seq', 18, true);


--
-- Name: pos_transaction_lines_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.pos_transaction_lines_id_seq', 228, true);


--
-- Name: pos_transactions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.pos_transactions_id_seq', 62, true);


--
-- Name: price_lists_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.price_lists_id_seq', 3, true);


--
-- Name: product_barcodes_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.product_barcodes_id_seq', 44, true);


--
-- Name: product_batches_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.product_batches_id_seq', 1, false);


--
-- Name: product_categories_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.product_categories_id_seq', 50, true);


--
-- Name: product_prices_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.product_prices_id_seq', 94, true);


--
-- Name: product_serial_numbers_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.product_serial_numbers_id_seq', 1, false);


--
-- Name: product_uom_conversions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.product_uom_conversions_id_seq', 87, true);


--
-- Name: product_variants_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.product_variants_id_seq', 1, true);


--
-- Name: products_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.products_id_seq', 79, true);


--
-- Name: profit_loss_analytics_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.profit_loss_analytics_id_seq', 1, false);


--
-- Name: promotions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.promotions_id_seq', 17, true);


--
-- Name: purchase_analytics_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.purchase_analytics_id_seq', 1, false);


--
-- Name: purchase_order_lines_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.purchase_order_lines_id_seq', 1, false);


--
-- Name: purchase_orders_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.purchase_orders_id_seq', 1, false);


--
-- Name: recipe_ingredients_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.recipe_ingredients_id_seq', 8, true);


--
-- Name: recipes_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.recipes_id_seq', 5, true);


--
-- Name: restaurant_order_items_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.restaurant_order_items_id_seq', 1, false);


--
-- Name: restaurant_orders_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.restaurant_orders_id_seq', 4, true);


--
-- Name: restaurant_tables_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.restaurant_tables_id_seq', 19, true);


--
-- Name: role_permissions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.role_permissions_id_seq', 323, true);


--
-- Name: role_ui_customizations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.role_ui_customizations_id_seq', 4, true);


--
-- Name: roles_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.roles_id_seq', 9, true);


--
-- Name: sales_analytics_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.sales_analytics_id_seq', 1, false);


--
-- Name: sales_order_lines_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.sales_order_lines_id_seq', 1, false);


--
-- Name: sales_orders_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.sales_orders_id_seq', 1, false);


--
-- Name: sales_return_lines_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.sales_return_lines_id_seq', 1, false);


--
-- Name: sales_returns_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.sales_returns_id_seq', 1, false);


--
-- Name: stock_count_lines_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.stock_count_lines_id_seq', 1, false);


--
-- Name: stock_counts_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.stock_counts_id_seq', 1, false);


--
-- Name: stock_movements_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.stock_movements_id_seq', 477, true);


--
-- Name: stock_reservations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.stock_reservations_id_seq', 1, false);


--
-- Name: storage_locations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.storage_locations_id_seq', 13, true);


--
-- Name: stores_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.stores_id_seq', 17, true);


--
-- Name: submenu_permissions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.submenu_permissions_id_seq', 145, true);


--
-- Name: submenus_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.submenus_id_seq', 133, true);


--
-- Name: suppliers_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.suppliers_id_seq', 1, false);


--
-- Name: tax_categories_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.tax_categories_id_seq', 3, true);


--
-- Name: transfer_request_items_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.transfer_request_items_id_seq', 21, true);


--
-- Name: transfer_requests_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.transfer_requests_id_seq', 21, true);


--
-- Name: ui_settings_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.ui_settings_id_seq', 5, true);


--
-- Name: units_of_measure_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.units_of_measure_id_seq', 20, true);


--
-- Name: uom_packaging_template_levels_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.uom_packaging_template_levels_id_seq', 15, true);


--
-- Name: uom_packaging_templates_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.uom_packaging_templates_id_seq', 5, true);


--
-- Name: user_roles_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.user_roles_id_seq', 30, true);


--
-- Name: user_store_access_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.user_store_access_id_seq', 49, true);


--
-- Name: users_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.users_id_seq', 44, true);


--
-- Name: waste_logs_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nembus_admin_user
--

SELECT pg_catalog.setval('public.waste_logs_id_seq', 1, false);


--
-- PostgreSQL database dump complete
--

\unrestrict c1HeHpRg3Idak5CAUN6FHDDG915sgtrinT5hLo7Gnzu9bDpfPfKEdpkEJwfTOii

