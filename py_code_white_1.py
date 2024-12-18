@classmethod
def get_descendants_ids(cls, dep_id):
        projection = {"_id": 1}
        ids = []
        for res in cls.get_descendants(dep_id, projection=projection):
            ids.append(res.get("_id"))
        return ids