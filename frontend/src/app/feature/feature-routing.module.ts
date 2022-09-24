import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {NewFeatureComponent} from "./new-feature/new-feature.component";
import {FeatureListComponent} from "./feature-list/feature-list.component";
import {EditFeatureComponent} from "./edit-feature/edit-feature.component";

const routes: Routes = [
  {path: '', component: FeatureListComponent},
  {path: 'new', component: NewFeatureComponent},
  {path: ':featureId/edit', component: EditFeatureComponent},
];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule]
})
export class FeatureRoutingModule {
}
