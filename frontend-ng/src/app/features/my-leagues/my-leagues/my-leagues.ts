import { ChangeDetectionStrategy, Component, computed, inject, OnInit, signal } from '@angular/core';
import { FormControl, FormGroup, ReactiveFormsModule } from '@angular/forms';
import { toSignal } from '@angular/core/rxjs-interop';
import { RouterLink } from '@angular/router';
import { TuiButton, TuiDataList, TuiDropdown, TuiGroup, TuiInput, TuiOption, TuiTextfield } from '@taiga-ui/core';
import { TuiBlock, TuiChevron } from '@taiga-ui/kit';
import { TuiSearch } from '@taiga-ui/layout';
import { MyLeagueStore } from '../my-league-store';
import { EnumLabelPipe } from '../../../shared/pipes/enum-label.pipe';
import { LeagueStatus } from '../../league/models/enums/league-status';
import { MyLeaguesList } from '../my-leagues-list/my-leagues-list';
import { PublicLeaguesList } from '../public-leagues-list/public-leagues-list';
import { Layout } from '../../../shared/components/layout/layout';

@Component({
  selector: 'app-my-leagues',
  imports: [
    Layout,
    EnumLabelPipe,
    ReactiveFormsModule,
    RouterLink,
    TuiBlock,
    TuiButton,
    TuiChevron,
    TuiDataList,
    TuiDropdown,
    TuiGroup,
    TuiInput,
    TuiOption,
    TuiSearch,
    TuiTextfield,
    MyLeaguesList,
    PublicLeaguesList,
  ],
  templateUrl: './my-leagues.html',
  styleUrl: './my-leagues.css',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class MyLeagues implements OnInit {
  protected readonly store = inject(MyLeagueStore);

  protected readonly statusOpen = signal(false);

  protected readonly form = new FormGroup({
    search: new FormControl(''),
    status: new FormControl<LeagueStatus | null>(null),
    tab: new FormControl<'myLeagues' | 'publicLeagues'>('myLeagues'),
  });

  private readonly formValue = toSignal(this.form.valueChanges, {
    initialValue: this.form.value,
  });

  protected readonly statusValues = Object.values(LeagueStatus).filter(
    (s) => s !== LeagueStatus.PENDING && s !== LeagueStatus.COMPLETED && s !== LeagueStatus.CANCELLED,
  );

  protected readonly filteredLeagues = computed(() => {
    const leagues = this.store.myLeagues();
    const { search } = this.formValue();
    const query = search?.toLowerCase() ?? '';

    return leagues.filter((league) => {
      return !query || league.Name.toLowerCase().includes(query);
    });
  });

  ngOnInit(): void {
    this.form.get('tab')!.valueChanges.subscribe((value) => {
      this.store.activeTab.set(value!);
    });
  }

  protected clearFilters(): void {
    this.form.get('search')!.setValue('');
    this.form.get('status')!.setValue(null);
  }

  protected selectStatus(status: LeagueStatus | null): void {
    this.form.get('status')!.setValue(status);
    this.statusOpen.set(false);
  }
}
