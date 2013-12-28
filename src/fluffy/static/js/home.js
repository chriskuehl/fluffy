var fileIndex = 0;
var fileInputs = [];

$(document).ready(function() {
	// browse button
	$("#selectFile").click(function() {
		input = addHiddenFileInput();
		input.trigger("click");
	});

	$("#selectFileInput").change(function() {
		console.log($(this).val());
	});

	// drop zone file drag-and-drop
	addDropZoneInput();
});

/**
 * Takes a file input element and handles displaying the file to the user and
 * storing the input to be submitted with the request.
 *
 * The input will be moved out of the existing parent and placed into a form
 * with all the other inputs being used for the upload.
 *
 * @param input - jQuery input object
 */
function handleInput(input) {
	console.log("Handling input: " + input.val());

	input.appendTo($("#uploadForm"));
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

/**
 * Create an input in #dropZone which can handle drag/drop file uploads
 */
function addDropZoneInput() {
	input = makeFileInput();

	input.appendTo($("#dropZone"));
	input.on("dragenter", dropZoneHoverOn);
	input.on("dragleave drop", dropZoneHoverOff);

	input.change(function() {
		addDropZoneInput();
	});
}

function dropZoneHoverOn() {
	$("#dropZone").css("backgroundColor", "#FFEECA");
}

function dropZoneHoverOff() {
	$("#dropZone").css("backgroundColor", "#F3FFE2");
}
