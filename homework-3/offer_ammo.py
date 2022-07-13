#!/usr/bin/env python3

import pandas as pd

def make_ammo(method, url, headers, case):
    req_template = (
          "%s %s HTTP/1.1\r\n"
          "%s\r\n"
          "\r\n"
    )
    req_template_w_entity_body = (
          "%s %s HTTP/1.1\r\n"
          "%s\r\n"
          "Content-Length: %d\r\n"
          "\r\n"
          "%s\r\n"
    )
    req = req_template % (method, url, headers)
    ammo_template = (
        "%d %s\n"
        "%s"
    )
    return ammo_template % (len(req), case, req)


def read_ids(file_name):
    ids = []
    with open(file_name) as fp:
        id = fp.readline()
        while id:
            ids.append(id.strip())
            id = fp.readline()
    return ids


def main():
    id_file = 'ids.csv'
    ammo_file = 'offer_ammo.txt'

    ids = read_ids(id_file)
    df = pd.DataFrame(ids, columns=['id'])

    top20 = df.sample(frac=0.2)
    top20['freq'] = 100
    top80 = df[~df.isin(top20)].dropna()
    top80['freq'] = 1

    df = pd.concat([top20, top80]).sample(n=1500000, replace=True, weights='freq')

    with open(ammo_file, 'a') as out:
        for index, row in df.iterrows():
            method, url, case = "GET", "/api/offer/%s/price" % row['id'], "offer_price"
            headers = "Host: hostname.com\r\n" + \
                "User-Agent: tank\r\n" + \
                "Accept: */*\r\n" + \
                "Connection: Close"
            out.write(make_ammo(method, url, headers, case))


if __name__ == "__main__":
    main()
