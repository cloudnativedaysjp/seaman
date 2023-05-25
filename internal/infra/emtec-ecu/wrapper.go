package emtec_ecu

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/cloudnativedaysjp/emtec-ecu/pkg/ws-proxy/schema"
)

type CndWrapper struct {
	Scene pb.SceneServiceClient
	Track pb.TrackServiceClient
}

func NewCndWrapper(scene pb.SceneServiceClient, track pb.TrackServiceClient) *CndWrapper {
	return &CndWrapper{scene, track}
}

//
// Scene
//

var _ pb.SceneServiceClient = (*CndWrapper)(nil)

func (w CndWrapper) ListScene(ctx context.Context, in *pb.ListSceneRequest, opts ...grpc.CallOption) (*pb.ListSceneResponse, error) {
	return w.Scene.ListScene(ctx, in, opts...)
}

func (w CndWrapper) MoveSceneToNext(ctx context.Context, in *pb.MoveSceneToNextRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return w.Scene.MoveSceneToNext(ctx, in, opts...)
}

//
// Track
//

var _ pb.TrackServiceClient = (*CndWrapper)(nil)

func (w CndWrapper) GetTrack(ctx context.Context, in *pb.GetTrackRequest, opts ...grpc.CallOption) (*pb.Track, error) {
	return w.Track.GetTrack(ctx, in, opts...)
}

func (w CndWrapper) ListTrack(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.ListTrackResponse, error) {
	return w.Track.ListTrack(ctx, in, opts...)
}

func (w CndWrapper) EnableAutomation(ctx context.Context, in *pb.SwitchAutomationRequest, opts ...grpc.CallOption) (*pb.Track, error) {
	return w.Track.EnableAutomation(ctx, in, opts...)
}

func (w CndWrapper) DisableAutomation(ctx context.Context, in *pb.SwitchAutomationRequest, opts ...grpc.CallOption) (*pb.Track, error) {
	return w.Track.DisableAutomation(ctx, in, opts...)
}
