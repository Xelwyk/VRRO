extends MeshInstance3D

var cam_mat = preload("res://materials/cam_material.tres")
@onready var camReq = $"../CamRequest"
var ready_to_request = true
# Called when the node enters the scene tree for the first time.
func _ready() -> void:
	camReq.request_completed.connect(self._cam_request_completed)


# Called every frame. 'delta' is the elapsed time since the previous frame.
func _process(delta: float) -> void:
	if ready_to_request:
		download_cam()
	
func updateCameramanPose(jsonPoints: String) -> void:
	if jsonPoints == null:
		return
		
	var point = JSON.parse_string(jsonPoints)
	
	if point == null:
		return
	
	var st = SurfaceTool.new()
	
	st.begin(Mesh.PRIMITIVE_POINTS)
	st.add_vertex(Vector3(point.x, point.y, point.z))
		
	self.mesh = st.commit()
	self.set_surface_override_material(0, cam_mat)


func download_cam() -> void:
	ready_to_request = false
	
	# Perform a GET request. The URL below returns JSON as of writing.
	var error = camReq.request("http://localhost:8081/camera")
	if error != OK:
		push_error("An error occurred in the HTTP request.")

func _cam_request_completed(result, response_code, headers, body):
	# Will print the user agent string used by the HTTPRequest node (as recognized by httpbin.org).
	updateCameramanPose(body.get_string_from_utf8())
	ready_to_request = true
