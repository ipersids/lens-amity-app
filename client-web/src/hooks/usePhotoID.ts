import { useQuery } from "@tanstack/react-query";
import { PhotoIDQueryKey } from "../constants";
import usersService from "../services/users";

const usePhotoID = (username?: string, photoID?: string) =>
  useQuery({
    queryKey: PhotoIDQueryKey(username ?? "", photoID ?? ""),
    queryFn: () => usersService.getPhotoByID(username ?? "", photoID ?? ""),
    enabled: !!username && !!photoID,
    staleTime: 60_000,
    retry: false,
  });

export default usePhotoID;
