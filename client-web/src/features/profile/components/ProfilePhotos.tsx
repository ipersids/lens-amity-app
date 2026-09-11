import { useInfiniteQuery } from "@tanstack/react-query";
import usersService from "../../../services/users";
import { PhotosFeed } from "../../photos";

type ProfilePhotosProps = {
  username: string;
  canViewPhotos: boolean;
};

const pageSize = 24;

const ProfilePhotos = ({ username, canViewPhotos }: ProfilePhotosProps) => {
  const { data, fetchNextPage, hasNextPage, isFetchingNextPage, isPending, isError } =
    useInfiniteQuery({
      queryKey: ["photos", username],
      queryFn: ({ pageParam }) => usersService.getUserPhotos(username, pageSize, pageParam),
      initialPageParam: undefined as string | undefined,
      getNextPageParam: (lastPage) => lastPage.nextCursor,
      enabled: canViewPhotos && !!username,
      staleTime: 60_000,
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

  const photos = data.pages.flatMap((page) => page.items);

  return (
    <>
      <PhotosFeed photos={photos} />
      {hasNextPage && (
        <button
          className="profile-photos-load-more"
          disabled={isFetchingNextPage}
          type="button"
          onClick={() => fetchNextPage()}
        >
          {isFetchingNextPage ? "Loading..." : "Load more"}
        </button>
      )}
    </>
  );
};

export default ProfilePhotos;
