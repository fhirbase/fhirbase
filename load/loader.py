import json
import logging
import os
import re
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
                    yield line
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
