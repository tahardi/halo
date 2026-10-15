# End-to-end flows

Maestro flows that run against the app on the iOS Simulator. CI runs them in the `app-e2e` job.

## Run locally

Requires a Mac with Xcode and [Maestro](https://maestro.mobile.dev) installed.

```bash
cd app
npx expo run:ios --configuration Release
maestro test e2e/
```

## Conventions

- Every flow lives in `app/e2e/*.yaml`.
- Every element a flow touches has a `testID` in the app code. Flows select it with `id:`.
