import json
import logging
import os
import random
import re
import time
import zipfile
import io


try:
    import requests
except ImportError:
    raise RuntimeError('Run `pip install requests` before using this tool')


logger = logging.getLogger()


def isndjson(filename):
    return filename.endswith('.ndjson')


def isjson(filename):
    return filename.endswith('.json')


def iszip(filename):
    return filename.endswith('.zip')


def dumpjson(res):
    return json.dumps(res, separators=(',', ':'))


def loadjson(jsonstr):
    return json.loads(jsonstr)


def prepare_json(jsonstr):
    """
    Processes Bundle or another resource and
    returns its ndjson representation as list
    """
    res = loadjson(jsonstr)

    resource_type = res.get('resourceType', None)
    if not resource_type:
        return []

    if resource_type == 'Bundle' and 'entry' in res:
        return [dumpjson(item['resource'])
                for item in res['entry']
                if 'resource' in item]
    else:
        return [dumpjson(res)]


def make_zip_openfn(zipfile):
    """
    Zipfile only opens file in binary mode but we want to work with text
    See https://stackoverflow.com/q/5627954
    """
    return lambda name, mode: io.TextIOWrapper(zipfile.open(name, mode))


def iter_lines_from_files(files, openfn=open):
    """
    Generator which yield lines for each file from `files`.
    Input files can have format .ndjson, .json or .zip
    """
    for filename in files:
        try:
            mode = 'rb' if iszip(filename) else 'r'
            with openfn(filename, mode) as fd:
                if isndjson(filename):
                    for line in fd:
                        yield line.rstrip('\n')
                elif isjson(filename):
                    for line in prepare_json(fd.read()):
                        yield line
                elif iszip(filename):
                    with zipfile.ZipFile(fd) as zf:
                        for line in iter_lines_from_files(
                                zf.namelist(), make_zip_openfn(zf)):
                            yield line
                else:
                    logger.warning(
                        'Unknown file format for {0}'.format(filename))
        except FileNotFoundError:
            logger.warning('Cannot open file {0}'.format(filename))


def iter_lines_from_external(url):
    with requests.get(url, stream=True) as resp:
        if resp.status_code == 200:
            if isndjson(url):
                for line in resp.iter_lines():
                    yield line.decode()
            elif isjson(url):
                for line in prepare_json(resp.content):
                    yield line
            elif iszip(url):
                buf = io.BytesIO(resp.content)
                with zipfile.ZipFile(buf) as zf:
                    for line in iter_lines_from_files(
                            zf.namelist(), make_zip_openfn(zf)):
                        yield line
            else:
                logger.warning('Unknown file format for {0}'.format(url))

        else:
            logger.warning('Cannot fetch url {0}'.format(url))


def iter_synthea_files(path):
    for filename in os.listdir(path):
        filepath = os.path.join(os.path.abspath(path), filename)
        if os.path.isfile(filepath) and filepath.endswith(".json"):
            yield filepath


def make_resource_ref(entrymap, url):
    if not url in entrymap:
        logger.warning("No URL {} resource mapping found".format(url))
        return url
    resource = entrymap[url]["resource"]
    fhirbase_ref = "{}/{}".format(resource["resourceType"], resource["id"])
    return fhirbase_ref


def make_fhirbase_ref(entrymap, url):
    return '"reference":"{}"'.format(make_resource_ref(entrymap, url))


def make_resource_entrymap(resources):
    entrymap = {}
    for entry in resources:
        fullurl, resource = entry.get("fullUrl"), entry.get("resource")
        if not fullurl or not resource:
            continue
        if "resourceType" in resource and "id" in resource:
            entrymap[fullurl] = entry
    return entrymap


def iter_flatten_synthea_resources(filename):
    with open(filename, "r") as bundle:
        resources = json.load(bundle).get("entry")
    if not resources:
        logger.warning("Skipping bundle {} with no resources".format(filename))
        return
    entrymap = make_resource_entrymap(resources)
    ref_regexp = re.compile(r'"reference":\s*?"(urn:uuid:.*?)"')
    for entry in resources:
        resource = entry.get("resource")
        if resource:
            yield re.sub(ref_regexp,
                    lambda ref: make_fhirbase_ref(entrymap, ref.group(1)),
                    json.dumps(resource, separators=(',', ':')))


def iter_lines_from_synthea(path):
    if not os.path.isdir(path):
        raise NotADirectoryError("Synthea FHIR folder expected")
    for filename in iter_synthea_files(path):
        for resource in iter_flatten_synthea_resources(filename):
            yield resource


def request_content_location(url):
    headers = {"Accept": "application/fhir+json",
               "Prefer": "respond-async"}
    resp = requests.get(url, headers=headers)
    if resp.status_code != requests.codes.accepted:
        return None
    return resp.headers.get("Content-Location")


def make_exp_backoff_seq(slot, ceiling):
    """
    Truncated Binary Exponential Backoff
    Generates a sequence of values in range 1 to (2^n - 1) multiplied by slot
    where n is in range 1 to ceiling (n increases with every yield).
    """
    random.seed()
    generation = 1
    while True:
        yield slot * random.randint(1, 2 ** generation - 1)
        if generation < ceiling:
            generation += 1


def poll_export_links_ready(url):
    resp = requests.get(url, headers={"Accept": "application/json"})
    progress = [requests.codes.accepted, requests.codes.too_many_requests]
    exp_backoff_timeout = make_exp_backoff_seq(10, 3)
    while resp.status_code in progress:
        retry_after = resp.headers.get("Retry-After")
        if retry_after and retry_after.isdigit():
            time.sleep(int(retry_after))
        else:
            time.sleep(next(exp_backoff_timeout))
        resp = requests.get(url, headers={"Accept": "application/json"})
    return resp


def decode_json_body(response):
    try:
        return response.json()
    except ValueError:
        return None


def get_export_links(url):
    resp = poll_export_links_ready(url)
    if resp.status_code != requests.codes.ok:
        return None
    if resp.headers.get("Content-Type") != "application/json":
        return None
    export_meta = decode_json_body(resp)
    if not export_meta:
        return None
    exported_resources = export_meta.get("output")
    if not exported_resources:
        return None
    return [res["url"] for res in exported_resources if "url" in res]


def iter_lines_from_bulk(url):
    content_location = request_content_location(url)
    if not content_location:
        logger.warning("Unable to get content location for {}".format(url))
        return
    links = get_export_links(content_location)
    if not links:
        logger.warning("Unable to get links to exported resources")
        return
    for link in links:
        for line in iter_lines_from_external(link):
            yield line
