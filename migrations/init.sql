--
-- PostgreSQL database dump
--

-- Dumped from database version 16.2
-- Dumped by pg_dump version 16.2

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;
--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: credentials; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.credentials (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    username text NOT NULL,
    password text NOT NULL,
    coin bigint DEFAULT 1000
);


ALTER TABLE public.credentials OWNER TO postgres;

--
-- Name: shops; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.shops (
    item text NOT NULL,
    price bigint NOT NULL
);


ALTER TABLE public.shops OWNER TO postgres;

--
-- Name: transactions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.transactions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    from_user uuid,
    to_user uuid,
    amount bigint
);


ALTER TABLE public.transactions OWNER TO postgres;

--
-- Name: user_items; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_items (
    user_id uuid,
    type text,
    quantity bigint
);


ALTER TABLE public.user_items OWNER TO postgres;

--
-- Data for Name: credentials; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.credentials (id, username, password, coin) FROM stdin;
6483741f-64d7-4cbf-8e01-37e31a1a850e	test	password	985
0eaef374-cfa0-4256-90bc-eb24c3200a79	test2	password	1005
\.


--
-- Data for Name: shops; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.shops (item, price) FROM stdin;
t-shirt	80
cup	20
book	50
pen	10
powerbank	200
hoody	300
umbrella	200
socks	10
wallet	50
pink-hoody	500
\.


--
-- Data for Name: transactions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.transactions (id, from_user, to_user, amount) FROM stdin;
14f9a618-6d1a-404a-b225-d7088c7f0b66	6483741f-64d7-4cbf-8e01-37e31a1a850e	0eaef374-cfa0-4256-90bc-eb24c3200a79	5
\.


--
-- Data for Name: user_items; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.user_items (user_id, type, quantity) FROM stdin;
6483741f-64d7-4cbf-8e01-37e31a1a850e	pen	1
\.


--
-- Name: credentials credentials_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.credentials
    ADD CONSTRAINT credentials_pkey PRIMARY KEY (id);


--
-- Name: shops shops_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.shops
    ADD CONSTRAINT shops_pkey PRIMARY KEY (item);


--
-- Name: transactions transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_pkey PRIMARY KEY (id);


--
-- Name: credentials uni_credentials_username; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.credentials
    ADD CONSTRAINT uni_credentials_username UNIQUE (username);


--
-- Name: idx_transaction_from_user; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_transaction_from_user ON public.transactions USING btree (from_user);


--
-- Name: idx_transaction_to_user; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_transaction_to_user ON public.transactions USING btree (to_user);


--
-- Name: idx_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_user_id ON public.credentials USING btree (id);


--
-- Name: idx_user_item_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_user_item_user_id ON public.user_items USING btree (user_id);


--
-- Name: idx_username; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_username ON public.credentials USING btree (username);


--
-- PostgreSQL database dump complete
--

