import { useQuery } from "@tanstack/react-query";
import usersService from "../../../services/users";
import { PhotosFeed } from "../../photos";

type ProfilePhotosProps = {
  username: string;
  canViewPhotos: boolean;
};

const ProfilePhotos = ({ username, canViewPhotos }: ProfilePhotosProps) => {
  const { data, isPending, isError } = useQuery({
    queryKey: ["photos", username],
    queryFn: () => usersService.getUserPhotos(username),
    enabled: canViewPhotos,
  });

  if (!canViewPhotos) {
    return <p>Can't see photos</p>;
  }

  if (isPending) {
    return <p>Loading...</p>;
  }

  if (isError) {
    return <p>Error</p>;
  }

  return <PhotosFeed photos={data.items} />;
};

export default ProfilePhotos;
