(function() {
	var sink = document.getElementById("stdstreams");
	var rID = sink.dataset.rid;
	var tID = sink.dataset.tid;
	var skip = sink.dataset.skip;

	if(skip == 0) {
		sink.innerHTML = "";
	}

	function get(url, callback) {
		var r = new XMLHttpRequest();
		r.open("GET", url);
		r.onreadystatechange = function() {
			if(r.readyState != 4) {
				return;
			}
			if(r.status != 200) {
				console.error("GET " + url + " with status " + r.status);
				return;
			}
			callback(r);
		}
		r.send();
	}

	function fetchStdStreams() {
		var url = "stdstreams?t=" + tID + "&r=" + rID + "&s=" + skip;
		get(url, function(req) {
			var json = JSON.parse(req.responseText);
			for(var i = 0; i < json.length; i++) {
				skip++;
				var line = json[i];
				var elem = document.createElement("div");
				elem.innerHTML = line.Text;
				elem.className = line.Stream == 1 ? "" : "stderr";
				sink.appendChild(elem);
			}

			if(req.getResponseHeader("Optask-Running") == "1") {
				setTimeout(fetchStdStreams, 200);
			} else {
				document.getElementById("running-indicator").remove();
				refreshStatus();
			}
		});
	};

	function refreshStatus() {
		var statusElem = document.getElementById("status");
		if(statusElem == null) {
			console.log("No element with ID 'status'");
			return;
		}

		var url = "status?t=" + tID + "&r=" + rID;
		get(url, function(req) {
			statusElem.innerHTML = req.responseText;
		});
	}

	fetchStdStreams();
})();
