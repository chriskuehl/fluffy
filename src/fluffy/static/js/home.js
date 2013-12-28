$(document).ready(function() {
	// browse button
	$("#selectFile").click(function() {
		$("#selectFileInput").trigger("click");
	});

	$("#selectFileInput").change(function() {
		console.log($(this).val());
	});


	// drop zone file drag-and-drop
	var dropZone = $("#dropZone");

	dropZone.on("dragenter", function(e) {
		e.preventDefault();
		e.stopPropagation();
	});

	dropZone.on("dragover", function(e) {
		$(this).css("backgroundColor", "#FFEECA");

		e.preventDefault();
		e.stopPropagation();
	});

	dropZone.on("dragleave", function() {
		$(this).css("backgroundColor", "#F3FFE2");
	});

	dropZone.on("drop", function(e) {
		e.preventDefault();

		alert("dropped.");
		return false;
	});
});
