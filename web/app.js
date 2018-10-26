window.onload = function() {
  var mime = "text/x-pgsql";

  const escapeHtml = unsafe => {
    return unsafe
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&#039;");
  };

  function tag(acc, tg, props) {
    acc.push("<" + tg);
    var cnt = Array.prototype.slice.call(arguments, 3);
    for (var k in props) {
      var v = props[k];
      acc.push(k + "=" + '"' + escapeHtml(v) + '"');
    }
    acc.push(">");
    Array.prototype.push.apply(acc, cnt);
    acc.push("</" + tg + ">");
  }

  const formatResultField = f => {
    if (typeof f === "object" && f !== null) {
      return "<pre>" + escapeHtml(JSON.stringify(f, null, 2)) + "</pre>";
    } else {
      return f;
    }
  };

  function runQuery(cm) {
    let q = cm.getValue();
    let url = new URL("/q", window.location);
    url.searchParams.append("query", q);

    try {
      gtag("event", "sql", {
        event_category: "fhirbase-demo",
        event_label: q
      });
    } catch (e) {
      console.error(e);
    }
    document.getElementById("results").innerHTML =
      "<center>Loading...</center>";

    fetch(url)
      .then(response => {
        return response
          .json()
          .then(json => Promise.resolve([response.status, json]));
      })
      .then(resp => {
        const status = resp[0];
        const json = resp[1];

        if (status === 200) {
          console.log("Got results", json);

          let tbl =
            '<h3>Results</h3><table class="table table-striped table-bordered table-sm"><thead><tr>';

          json.columns.forEach(clmn => {
            tbl += "<th>" + clmn.Name + "</th>";
          });

          tbl += "</tr></thead><tbody>";

          json.rows.forEach(row => {
            tbl +=
              "<tr>" +
              row.map(f => "<td>" + formatResultField(f) + "</td>").join("") +
              "</tr>";
          });

          tbl += "</tbody></table>";

          document.getElementById("results").innerHTML = tbl;
        } else {
          document.getElementById("results").innerHTML =
            "<h3>Results</h3><div class='alert alert-danger'>" +
            json.message +
            "</div>";
        }
      })
      .catch(err => {
        document.getElementById("results").innerHTML =
          "<h3>Results</h3><div class='alert alert-danger'>" +
          err.message +
          "</div>";
      });
    return false;
  }

  window.submitQuery = () => {
    runQuery(window.editor);
  };

  window.editor = CodeMirror(document.getElementById("editor"), {
    mode: mime,
    theme: "duotone-light",
    indentWithTabs: true,
    smartIndent: true,
    lineNumbers: true,
    matchBrackets: true,
    value: "SELECT * FROM patient LIMIT 100;",
    autofocus: true,
    extraKeys: {
      "Ctrl-Space": "autocomplete",
      "Ctrl-Enter": runQuery
    },
    hintOptions: {
      tables: {
        users: ["name", "score", "birthDate"],
        countries: ["name", "population", "size"]
      }
    }
  });

  var data = {};
  window.doSelect = idx => {
    var item = data.queries[idx];
    item && item.query && window.editor.setValue(item.query);
  };

  window.doExec = idx => {
    var item = data.queries[idx];
    item && item.query && window.editor.setValue(item.query);
    runQuery(window.editor);
  };

  fetch("https://fhirbase.github.io/demo-data/default.json")
    .then(response => {
      return response
        .json()
        .then(json => Promise.resolve({ status: response.status, data: json }));
    })
    .then(resp => {
      var res = [];
      data = resp.data;
      tag(res, "h3", {}, data.title);
      data.queries.forEach((x, i) => {
        tag(
          res,
          "a",
          {
            class: "item",
            href: "javascript:void(0)",
            title: x.query,
            onClick: "doSelect(" + i + ")",
            ondblclick: "doExec(" + i + ")"
          },
          x.title || x.query
        );
      });
      document.getElementById("right").innerHTML = res.join(" ");
    });
};
