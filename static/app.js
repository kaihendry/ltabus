function countdown(id, arrival) {
  if (!id) {
    return;
  }
  // Recompute from the clock on every tick. A frozen tab - a locked phone,
  // a backgrounded tab - fires no timers, so decrementing a captured value
  // would leave a stale time on screen and drift slow.
  var seconds = (arrival - Date.now()) / 1000;
  if (Math.abs(seconds) > 60) {
    id.textContent = Math.trunc(seconds / 60) + "m";
  } else {
    id.textContent = Math.trunc(seconds) + "s";
  }
  setTimeout(countdown, 1000, id, arrival);
}

// localStorage throws outright when site data is blocked, and the stored
// value can be corrupt. Neither may take out the rest of the load handler,
// which is also what redirects to the closest bus stop.
function readHistory() {
  try {
    return JSON.parse(window.localStorage.getItem("history")) || {};
  } catch (e) {
    console.warn("discarding unreadable history", e);
    return {};
  }
}

function writeHistory(history) {
  try {
    window.localStorage.setItem("history", JSON.stringify(history));
  } catch (e) {
    console.warn("history not saved", e);
  }
}

window.addEventListener("load", function () {
  var timings = document.getElementsByTagName("time");
  for (let i = 0; i < timings.length; i++) {
    var arrival = new Date(timings[i].getAttribute("datetime"));
    countdown(timings[i], arrival.getTime());
  }
  countdown(document.getElementById("lastupdated"), Date.now());

  var history = readHistory();
  var busstopcode = document.getElementById("id")?.value;

  if (busstopcode) {
    var stop = history[busstopcode] || { count: 0 };
    stop.count++;
    stop.name = document.getElementById("namedBusStop")?.textContent || "";
    history[busstopcode] = stop;
    writeHistory(history);
  } else if (navigator.geolocation) {
    navigator.geolocation.getCurrentPosition(function (position) {
      window.location =
        "/closest?lat=" +
        position.coords.latitude +
        "&lng=" +
        position.coords.longitude;
    });
  }

  var stations = document.getElementById("stations");
  Object.keys(history)
    .sort(function (a, b) {
      return history[b].count - history[a].count;
    })
    .forEach(function (code) {
      var link = document.createElement("a");
      link.href = "/?id=" + code;
      link.textContent = history[code].name
        ? code + " " + history[code].name + " (" + history[code].count + ")"
        : code;
      var li = document.createElement("li");
      li.appendChild(link);
      stations.appendChild(li);
    });
});
