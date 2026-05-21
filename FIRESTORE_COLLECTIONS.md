# Firestore Collections - PescaApp

## Colecciones a crear en Firebase Firestore

### 1. `users`
```
Document ID: auto-generated
Fields:
  - email: string (UNIQUE - verificar en código)
  - name: string
  - photoUrl: string | null
  - role: string ("user" | "admin")
  - createdAt: timestamp
  - updatedAt: timestamp
```

### 2. `species`
```
Document ID: auto-generated
Fields:
  - commonName: string
  - scientificName: string (UNIQUE - verificar en código)
  - description: string
  - imageUrl: string | null
  - habitat: string
  - diet: string
  - averageSizeCm: number
  - averageWeightKg: number
  - maxSizeCm: number
  - maxWeightKg: number
  - fishingTips: array<string>
  - createdAt: timestamp
  - updatedAt: timestamp
```

### 3. `fishing_spots`
```
Document ID: auto-generated
Fields:
  - name: string
  - description: string
  - latitude: number
  - longitude: number
  - region: string
  - waterType: string ("river" | "lake" | "sea" | "lagoon")
  - boatAllowed: boolean
  - boatRequired: boolean
  - access: string ("easy" | "moderate" | "difficult" | "boatOnly")
  - createdByUserId: string (reference to users)
  - views: number (default: 0)
  - averageRating: number (default: 0)
  - totalRatings: number (default: 0)
  - totalComments: number (default: 0)
  - createdAt: timestamp
  - updatedAt: timestamp
```

### 4. `spot_species`
```
Document ID: auto-generated
Fields:
  - spotId: string (reference to fishing_spots)
  - speciesId: string (reference to species)
  - recommendedBaits: array<string>
  - recommendedRod: string
  - recommendedLine: string
  - recommendedHook: string
  - bestSeason: string
  - difficulty: string ("easy" | "medium" | "hard")
  - notes: string
```

### 5. `comments`
```
Document ID: auto-generated
Fields:
  - spotId: string (reference to fishing_spots)
  - userId: string (reference to users)
  - text: string
  - likes: number (default: 0)
  - createdAt: timestamp
  - updatedAt: timestamp | null
  - deletedAt: timestamp | null (soft delete)
```

### 6. `comment_likes`
```
Document ID: auto-generated
Fields:
  - commentId: string (reference to comments)
  - userId: string (reference to users)
  - createdAt: timestamp

Composite Index: commentId + userId (UNIQUE - verificar en código)
```

### 7. `ratings`
```
Document ID: auto-generated
Fields:
  - spotId: string (reference to fishing_spots)
  - userId: string (reference to users)
  - stars: number (1-5)
  - createdAt: timestamp
  - updatedAt: timestamp

Composite Index: spotId + userId (UNIQUE - verificar en código)
```

### 8. `favorites`
```
Document ID: auto-generated
Fields:
  - userId: string (reference to users)
  - spotId: string (reference to fishing_spots)
  - createdAt: timestamp

Composite Index: userId + spotId (UNIQUE - verificar en código)
```

---

## Índices Compuestos Requeridos en Firestore

Crear estos índices en la consola de Firebase o via `firestore.indexes.json`:

1. **comments**: `spotId ASC, createdAt DESC`
2. **comments**: `spotId ASC, likes DESC`
3. **ratings**: `spotId ASC, createdAt DESC`
4. **ratings**: `spotId ASC, userId ASC` (para unicidad)
5. **favorites**: `userId ASC, createdAt DESC`
6. **favorites**: `userId ASC, spotId ASC` (para unicidad)
7. **fishing_spots**: `region ASC, averageRating DESC`
8. **fishing_spots**: `waterType ASC, averageRating DESC`
9. **fishing_spots**: `createdAt DESC`
10. **comment_likes**: `commentId ASC, userId ASC` (para unicidad)
11. **spot_species**: `spotId ASC`
12. **spot_species**: `speciesId ASC`

## Reglas de Seguridad

Las reglas de seguridad se manejan a nivel de aplicación (backend), no en Firestore rules,
ya que el acceso es server-side con service account.

```
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {
    match /{document=**} {
      allow read, write: if false; // Solo acceso desde backend con service account
    }
  }
}
```

