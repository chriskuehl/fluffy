var fileIndex = 0;
var allFiles = [];

var uploading = false;
var uploadCompleted = false;
var uploadRequest;
var uploadSamples = [];

var transitioningModes = false;

$(document).ready(function() {
    $('.switch-modes a').click(function() {
        if (transitioningModes) {
            return;
        }

        transitioningModes = true;
        var duration = 200;
        var pb = $('.pastebinForm'),
            fh = $('#file-holder'),
            container = $('#container');

        if (fh.is(':visible')) {
            container.animate({'width': '960px'}, duration);
            fh.slideUp(duration);
            pb.slideDown(duration, function() {
                transitioningModes = false;
            });
        } else {
            container.animate({'width': '580px'}, duration);
            pb.slideUp(duration);
            fh.slideDown(duration, function() {
                transitioningModes = false;
            });
        }
    });

    // browse button
    $("#selectFile").click(function() {
        input = addHiddenFileInput();
        input.trigger("click");
    });

    $("#selectFileInput").change(function() {
        console.log($(this).val());
    });

    // drop zone
    var dropZone = $("#dropZone");
    dropZone.on("dragover dragenter", dropZoneHoverOn);
    dropZone.on("dragleave drop", dropZoneHoverOff);
    dropZone.on("drop", function(e) {
        var oe = e.originalEvent;

        if (oe.dataTransfer) {
            var dt = oe.dataTransfer;

            for (var i = 0; i < dt.files.length; i ++) {
                queueFile(dt.files[i]);
            }
        } else {
            alert("Sorry, your browser doesn't seem to support file dropping.");
        }
    });

    // uploading
    $("#upload").click(function() {
        if (! uploading) {
            upload();
        } else if (! uploadCompleted) {
            cancelUpload();
        }
    });
});

/**
 * Upload the queued files via XHR. Takes care of updating the UI (displaying
 * progress, hiding remove buttons, etc.)
 */
function upload() {
    if (! canUpload()) {
        return alert("Can't upload right now.");
    }

    uploading = true;

    // update UI
    var uploadButton = $("#upload");
    uploadButton.text("Cancel Upload");

    $("#statusText").show();
    $("#selectFiles").slideUp(200);
    $(".remove").fadeOut(200);

    // start the upload
    // http://stackoverflow.com/a/8244082
    var request = $.ajax({
        url: "/upload?json",
        type: "POST",
        contentType: false,

        data: getFormData(),
        processData: false,

        xhr: function() {
            var req = $.ajaxSettings.xhr();

            req.upload.addEventListener("progress", function(e) {
                if (request != uploadRequest) {
                    return; // upload was cancelled
                }

                if (e.lengthComputable) {
                    updateProgress(e.loaded, e.total);
                }
            }, false);

            return req;
        },

        error: function(xhr, status) {
            if (request !== uploadRequest) {
                return; // upload was cancelled
            }

            if (xhr.status === 413) {
                alert('Sorry! This request exceeds the maximum upload size.');
                cancelUpload();
            } else {
                // TODO: improve error handling
                console.log("Unhandled failure: " + status + ", status=" + xhr.status + ", statusText=" + xhr.statusText);
                alert("Sorry, an unexpected error occured.");
            }
        },

        success: function(data) {
            if (request != uploadRequest) {
                return; // upload was cancelled
            }

            // TODO: improve error handling
            if (! data.success) {
                return alert(data.error);
            }

            window.location.href = data.redirect;
        }
    });

    uploadRequest = request;
}

/**
 * Cancels the current upload and restores the UI back to pre-upload state.
 */
function cancelUpload() {
    oldRequest = uploadRequest;

    uploading = false;
    uploadRequest = null;
    uploadSamples = [];

    if (oldRequest) {
        oldRequest.abort();
    }

    // update UI
    var uploadButton = $("#upload");
    uploadButton.text("Start Upload");

    $("#statusText").hide();
    $("#selectFiles").slideDown(200);
    $(".remove").fadeIn(200);

    $(".progress").css("width", "0%");
}

/**
 * Updates progress based on the number of bytes uploaded.
 *
 * Progress is displayed per-file even though we only have the current number
 * of bytes uploaded and the total to be uploaded (and don't actually know
 * which files have been or are being uploaded). It looks better, though, and
 * still gives an accurate representation of overall progress.
 *
 * @param bytes
 * @param totalBytes
 */
function updateProgress(bytes, totalBytes) {
    if (uploadCompleted) {
        return;
    }

    if (bytes >= totalBytes) {
        uploadCompleted = true;

        // hide uploading UI, show loading orb
        $("#file-holder").children(":visible").animate({
            opacity: 0
        }, 350);

        var statusText = "Upload complete, storing file" +
            plural(allFiles.length) + "...";
        $("#loading > p").text(statusText);
        $("#loading").fadeIn(350);

        return;
    }

    // progress bars on individual files
    var bytesLeft = bytes;
    var fileList = $("#files");

    fileList.children().each(function() {
        var row = $(this);
        var size = row.data("file").size;
        var progress = 0;

        if (bytesLeft > 0) {
            progress = Math.min(1, bytesLeft / size);
            bytesLeft -= size;
        }

        var progressInt = Math.floor(100 * progress);
        row.find(".progress").css("width", progressInt + "%");
    });

    // status text
    var cur = getHumanSize(bytes);
    var total  = getHumanSize(totalBytes);
    var percent = Math.floor((bytes / totalBytes) * 100);

    if (isNaN(percent) || percent < 0 || percent > 100) {
        percent = 0;
    }

    var uploadRate = calculateUploadRate(bytes);
    var bull = String.fromCharCode(8226); // bullet character

    var line1 = cur + " / " + total + " (" + percent + "%)";
    var line2 = "";

    if (uploadRate) {
        var humanUploadRate = getHumanSize(uploadRate) + "/s";
        var secondsRemaining = (totalBytes - bytes) / uploadRate;
        var timeRemaining = getHumanTime(secondsRemaining);

        line2 = humanUploadRate + " " + bull + " " + timeRemaining + " remaining"
    }

    $("#statusText").html(htmlEncode(line1) + "<br />" + htmlEncode(line2));
}

/**
 * Estimate the current upload rate based on history of progress snapshots
 * collected in the past SAMPLE_PERIOD milliseconds.
 *
 * @param bytes - current number of bytes uploaded
 * @return upload rate in bytes/sec OR null (if we can't estimate yet)
 */
var SAMPLE_PERIOD = 15 * 1000; // time to keep samples, in milliseconds
var REQUIRED_SAMPLES = 5; // # of samples required to make an estimate

function calculateUploadRate(bytes) {
    var now = new Date().getTime();
    uploadSamples.push([bytes, now]);

    // get rid of old samples
    while (uploadSamples[0][1] < (now - SAMPLE_PERIOD)) {
        uploadSamples.shift();
    }

    if (uploadSamples.length < REQUIRED_SAMPLES) {
        return null;
    }

    return 1000 * ((bytes - uploadSamples[0][0]) / (now - uploadSamples[0][1]));
}

/**
 * Encode text for HTML.
 *
 * Source: http://stackoverflow.com/a/1219983
 *
 * @param value
 * @return encoded text
 */
function htmlEncode(value) {
    return $("<div />").text(value).html();
}

/**
 * Return "s" when some quantity should be plural.
 *
 * @param count
 * @return "s" or empty string
 */
function plural(count) {
    return count == 1 ? "" : "s";
}

/**
 * Convert a seconds count into a human-readable time string like
 * "3 minutes, 7 seconds".
 *
 * For readability, only one lower unit is used, i.e. you can have "X minutes,
 * Y seconds" or "X hours, Y minutes", but never "X hours, Y minutes, Z
 * seconds".
 *
 * @param seconds
 * @return human-readable time
 */
var ONE_HOUR = 3600;
var ONE_MINUTE = 60;

function getHumanTime(seconds) {
    var units = ["hour", "minute", "second"];
    var times = [
        Math.floor(seconds / ONE_HOUR),
        Math.floor(seconds / ONE_MINUTE) % 60,
        Math.floor(seconds) % 60
    ];

    // cut off any zero times at the start
    while (times.length > 1 && times[0] == 0) {
        units.shift();
        times.shift();
    }

    var str = "";

    for (var i = 0; i < Math.min(2, times.length); i ++) {
        var time = times[i];
        var unit = units[i];

        str += time + " " + unit + plural(time) + ", ";
    }

    return str.substring(0, str.length - 2);
}

/**
 * Convert a byte count into a human-readable size string like "4.2 MB".
 *
 * Roughly based on Apache Commons FileUtils#byteCountToDisplaySize:
 * https://commons.apache.org/proper/commons-io/
 *
 * @param bytes
 * @return human-readable size
 */
var ONE_GB = 1073741824;
var ONE_MB = 1048576;
var ONE_KB = 1024;

function getHumanSize(size) {
    if (size / ONE_GB >= 1) {
        return (size / ONE_GB).toFixed(1) + " GB";
    } else if (size / ONE_MB >= 1) {
        return (size / ONE_MB).toFixed(1) + " MB";
    } else if (size / ONE_KB >= 1) {
        return (size / ONE_KB).toFixed(1) + " KB";
    } else {
        return size + " bytes";
    }
}

/**
 * @return FormData object containing files to be uploaded
 */
function getFormData() {
    var formData = new FormData();
    for (var i = 0; i < allFiles.length; i ++) {
        formData.append("file", allFiles[i]);
    }
    return formData;
}

/**
 * Takes a file input element and handles displaying the file to the user and
 * storing the file to be submitted with the request.
 *
 * The input will be removed from the DOM.
 *
 * @param input - jQuery input object
 */
function handleInput(input) {
    if (uploading) {
        return;
    }

    var files = input[0].files;

    for (var i = 0; i < files.length; i ++) {
        queueFile(files[i]);
    }

    input.remove();
}

/**
 * Queues a file by pushing it to the allFiles array and displaying it in
 * the file list.
 */
function queueFile(file) {
    if (! fileAlreadyQueued(file)) {
        allFiles.push(file);
        displayFile(file);
    }
}

/**
 * Checks if a file is already queued to be uploaded, checking for duplicates
 * based on name.
 *
 * @return whether or not a file is already queued to be uploaded
 */
function fileAlreadyQueued(file) {
    for (var i = 0; i < allFiles.length; i ++) {
        if (allFiles[i].name == file.name) {
            return true;
        }
    }

    return false;
}

/**
 * Displays a file in the file list.
 *
 * @param file - File object
 */
function displayFile(file) {
    var li = $("<li />");
    li.data("file", file);

    var progress = $("<div />");
    progress.addClass("progress");
    progress.appendTo(li);

    var icon = $("<img />");
    icon.attr("src", getIcon(file.name));
    icon.appendTo(li);

    var title = $("<div />");
    title.addClass("title");
    title.text(file.name);
    title.appendTo(li);

    var remove = $("<a />");
    remove.addClass("remove");
    remove.html("&times;");
    remove.appendTo(li);

    remove.click(function() {
        if (uploading) {
            return;
        }

        var idx = allFiles.indexOf(file);

        if (idx > -1) {
            allFiles.splice(idx, 1);
        }

        li.remove();
        updateUpload();
    });

    li.appendTo($("#files"));
    updateUpload();
}

/**
 * Show or hide the upload button based on whether or not any files are
 * queued for upload.
 */
function updateUpload() {
    var visible = $("#upload").is(":visible");
    var shouldBeVisible = canUpload();

    if (visible && ! shouldBeVisible) {
        $("#upload").hide();
    } else if (! visible && shouldBeVisible) {
        // set display -> block instead of calling show() since otherwise
        // jQuery will "restore" it to inline
        $("#upload").css("display", "block");
    }
}

/**
 * @return whether or not an upload can proceed
 */
function canUpload() {
    return allFiles.length > 0;
}

/**
 * Returns the path of the icon for a file.
 *
 * @param fileName
 * @return relative path to the icon
 */
function getIcon(fileName) {
    var parts = fileName.split(".");
    var extension = parts[parts.length - 1].toLowerCase();
    return icons[extension] || icons['unknown'];
}

/**
 * Creates a new hidden file input in <body>. You probably want to call click()
 * on the return value to display the browse window.
 *
 * @return jQuery input object
 */
function addHiddenFileInput() {
    var input = makeFileInput();

    input.appendTo($("body"));
    input.addClass("hiddenFileInput");

    return input;
}

/**
 * Makes a new file input with a unique name which will call handleInput (and
 * thus take care of everything needed for files to be uploaded).
 */
function makeFileInput() {
    var input = $("<input type=\"file\" />");

    input.attr({
        name: "file-" + (fileIndex ++) + "[]",
        multiple: "multiple"
    });

    input.change(function() {
        handleInput(input);
    });

    return input;
}

function dropZoneHoverOn(e) {
    $("#dropZone").css("backgroundColor", "#FFEECA");

    e.preventDefault();
    return false;
}

function dropZoneHoverOff(e) {
    $("#dropZone").css("backgroundColor", "#F3FFE2");

    e.preventDefault();
    return false;
}
